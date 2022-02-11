// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// http://www.eclipse.org/legal/epl-2.0
//
// SPDX-License-Identifier: EPL-2.0

package jsonfile

import (
	"bytes"
	"time"
	"unicode/utf8"

	"github.com/eclipse-kanto/container-management/containerm/logger"
)

const (
	timeLayout = time.RFC3339Nano
	hexEncode  = "0123456789abcdef"

	symbolEscape                    = '\\'
	symbolDoubleQuotes              = '"'
	symbolNewLine                   = '\n'
	symbolTab                       = '\t'
	symbolReturn                    = '\r'
	symbolGreaterThan               = '>'
	symbolLessThan                  = '<'
	symbolAmpersand                 = '&'
	symbolUnicodeLineSeparator      = '\u2028'
	symbolUnicodeParagraphSeparator = '\u2029'
)

// type jsonLogEntry struct {
// 	Stream     string            `json:"stream,omitempty"`
// 	Log        string            `json:"log,omitempty"`
// 	Timestamp  time.Time         `json:"time"`
// 	Attributes map[string]string `json:"attributes,omitempty"`
// }

// performUnmarshal parses the provided io.Reader to the LogMessage
// func performUnmarshal(r io.Reader) func() (*logger.LogMessage, error) {
// 	dec := json.NewDecoder(r)

// 	return func() (*logger.LogMessage, error) {
// 		logEntry := &jsonLogEntry{}
// 		if err := dec.Decode(logEntry); err != nil {
// 			return nil, err
// 		}

// 		return &logger.LogMessage{
// 			Source:     logEntry.Stream,
// 			Line:       []byte(logEntry.Log),
// 			Timestamp:  logEntry.Timestamp,
// 			Attributes: logEntry.Attributes,
// 		}, nil
// 	}
// }

func marshalLogMessageToJSONBytes(message *logger.LogMessage) ([]byte, error) {
	var (
		isFirst = true
		buffer  bytes.Buffer
	)

	buffer.WriteString("{")
	if len(message.Source) != 0 {
		isFirst = false
		buffer.WriteString(`"stream":`)
		bytesToJSONString(&buffer, []byte(message.Source))
	}

	if len(message.Line) != 0 {
		if !isFirst {
			buffer.WriteString(`,`)
		}
		isFirst = false
		buffer.WriteString(`"log":`)
		bytesToJSONString(&buffer, message.Line)
	}

	if !isFirst {
		buffer.WriteString(`,`)
	}

	buffer.WriteString(`"time":`)
	buffer.WriteString(message.Timestamp.UTC().Format(`"` + timeLayout + `"`))

	buffer.WriteString(`}`)

	// NOTE: add newline here to make the decoder easier
	buffer.WriteByte('\n')

	bs := buffer.Bytes()
	buffer.Reset()
	return bs, nil
}

var specialSymbols = []byte{symbolEscape, symbolDoubleQuotes, symbolLessThan, symbolGreaterThan, symbolAmpersand}

func checkIsSpecial(b byte) bool {
	for _, specialSymbol := range specialSymbols {
		if b == specialSymbol {
			return true
		}
	}
	return false
}

// bytesToJSONString copies from encoding/json/encode.go#stringBytes
func bytesToJSONString(buffer *bytes.Buffer, bytes []byte) {

	buffer.WriteByte(symbolDoubleQuotes)
	start := 0
	for i := 0; i < len(bytes); {
		if currentByte := bytes[i]; currentByte < utf8.RuneSelf {
			if 0x20 <= currentByte && !checkIsSpecial(currentByte) {
				i++
				continue

			}

			if start < i {
				buffer.Write(bytes[start:i])
			}
			switch currentByte {
			case symbolEscape, symbolDoubleQuotes:
				buffer.WriteByte(symbolEscape)
				buffer.WriteByte(currentByte)
			case symbolNewLine:
				buffer.WriteByte(symbolEscape)
				buffer.WriteByte('n')
			case symbolReturn:
				buffer.WriteByte(symbolEscape)
				buffer.WriteByte('r')
			case symbolTab:
				buffer.WriteByte(symbolEscape)
				buffer.WriteByte('t')
			default:
				buffer.WriteString(`\u00`)
				buffer.WriteByte(hexEncode[currentByte>>4])
				buffer.WriteByte(hexEncode[currentByte&0xF])
			}
			i++
			start = i
			continue
		}

		c, size := utf8.DecodeRune(bytes[i:])
		if c == utf8.RuneError && size == 1 {
			if start < i {
				buffer.Write(bytes[start:i])
			}
			buffer.WriteString(`\ufffd`)
			i += size
			start = i
			continue
		}

		if c == symbolUnicodeLineSeparator || c == symbolUnicodeParagraphSeparator {
			if start < i {
				buffer.Write(bytes[start:i])
			}
			buffer.WriteString(`\u202`)
			buffer.WriteByte(hexEncode[c&0xF])
			i += size
			start = i
			continue
		}
		i += size
	}
	if start < len(bytes) {
		buffer.Write(bytes[start:])
	}
	buffer.WriteByte(symbolDoubleQuotes)
}
