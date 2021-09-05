// Copyright 2021 coord_e
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//  	 http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package exporter

import (
	"fmt"
	"os"
)

func listAllAdapters() ([]uint, error) {
	files, err := os.ReadDir("/dev/dvb")
	if err != nil {
		return nil, err
	}

	var adapters []uint
	for _, file := range files {
		if !file.IsDir() {
			continue
		}
		var idx uint
		if _, err := fmt.Sscanf(file.Name(), "adapter%d", &idx); err != nil {
			continue
		}
		adapters = append(adapters, idx)
	}

	return adapters, nil
}

func listAllFrontends(adapter uint) ([]uint, error) {
	dev := fmt.Sprintf("/dev/dvb/adapter%d", adapter)
	files, err := os.ReadDir(dev)
	if err != nil {
		return nil, err
	}

	var frontends []uint
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		var idx uint
		if _, err := fmt.Sscanf(file.Name(), "frontend%d", &idx); err != nil {
			continue
		}
		frontends = append(frontends, idx)
	}

	return frontends, nil
}
