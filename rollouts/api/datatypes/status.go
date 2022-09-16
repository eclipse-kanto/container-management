// Copyright (c) 2021 Contributors to the Eclipse Foundation
//
// See the NOTICE file(s) distributed with this work for additional
// information regarding copyright ownership.
//
// This program and the accompanying materials are made available under the
// terms of the Eclipse Public License 2.0 which is available at
// https://www.eclipse.org/legal/epl-2.0, or the Apache License, Version 2.0
// which is available at https://www.apache.org/licenses/LICENSE-2.0.
//
// SPDX-License-Identifier: EPL-2.0 OR Apache-2.0

package datatypes

// Status represents the Status Vorto SUv2 datatype
type Status string

const (
	// Started operation status
	Started Status = "STARTED"
	// Downloading operation status
	Downloading Status = "DOWNLOADING"
	// DownloadingWaiting operation status
	DownloadingWaiting Status = "DOWNLOADING_WAITING"
	// Downloaded operation status
	Downloaded Status = "DOWNLOADED"
	// Installing operation status
	Installing Status = "INSTALLING"
	// InstallingWaiting operation status
	InstallingWaiting Status = "INSTALLING_WAITING"
	// Installed operation status
	Installed Status = "INSTALLED"
	// Removing operation status
	Removing Status = "REMOVING"
	// RemovingWaiting operation status
	RemovingWaiting Status = "REMOVING_WAITING"
	// Removed operation status
	Removed Status = "REMOVED"
	// Canceling operation status
	Canceling Status = "CANCELING"
	// CancelingWaiting operation status
	CancelingWaiting Status = "CANCELING_WAITING"
	// CancelRejected operation status
	CancelRejected Status = "CANCEL_REJECTED"
	// FinishedCanceled operation status
	FinishedCanceled Status = "FINISHED_CANCELED"
	// FinishedError operation status
	FinishedError Status = "FINISHED_ERROR"
	// FinishedSuccess operation status
	FinishedSuccess Status = "FINISHED_SUCCESS"
	// FinishedWarning operation status
	FinishedWarning Status = "FINISHED_WARNING"
	// FinishedRejected operation status
	FinishedRejected Status = "FINISHED_REJECTED"
)
