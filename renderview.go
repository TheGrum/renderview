// Copyright 2016 Howard C. Shaw III. All rights reserved.
// Use of this source code is governed by the MIT-license
// as defined in the LICENSE file.

package renderview

const (
	OPT_NONE        = iota      // 0
	OPT_CENTER_ZOOM = 1 << iota // 1
	OPT_AUTO_ZOOM   = 1 << iota // 2
)

const ZOOM_RATE = 0.1
