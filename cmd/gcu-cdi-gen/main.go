/*
 * Copyright 2026 The Kubernetes Authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific license governing permissions and
 * limitations under the License.
 */

// Command gcu-cdi-gen writes Enflame GCU CDI specs into a directory for
// comparison against enflame-ctk output on a real node.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	gcucdi "github.com/Project-HAMi/enflame-dra-driver/pkg/cdi"
)

func main() {
	var (
		outDir     = flag.String("out", "", "directory to write CDI specs (required)")
		hostRoot   = flag.String("host-root", "/", "host filesystem root used for dependency discovery")
		devRoot    = flag.String("dev-root", "/dev", "device node root")
		driverName = flag.String("driver-name", "gcu.enflame.com", "DRA driver name")
		class      = flag.String("class", "gcu", "CDI class")
		nodeName   = flag.String("node-name", os.Getenv("NODE_NAME"), "node name injected into base spec env")
		indices    = flag.String("indices", "", "comma-separated GCU indices for a sample claim spec (optional)")
		videoNodes = flag.Bool("video-nodes", false, "include /dev/gcuNvid* nodes in the sample claim spec")
		claimUID   = flag.String("claim-uid", "00000000-0000-0000-0000-000000000001", "claim UID used for the sample claim spec")
	)
	flag.Parse()

	if *outDir == "" {
		fmt.Fprintln(os.Stderr, "--out is required")
		os.Exit(2)
	}
	if err := os.MkdirAll(*outDir, 0o750); err != nil {
		fmt.Fprintf(os.Stderr, "create out dir: %v\n", err)
		os.Exit(1)
	}

	mgr, err := gcucdi.NewManager(
		gcucdi.WithCDIRoot(*outDir),
		gcucdi.WithHostRoot(*hostRoot),
		gcucdi.WithDevRoot(*devRoot),
		gcucdi.WithDriverName(*driverName),
		gcucdi.WithClass(*class),
		gcucdi.WithNodeName(*nodeName),
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create CDI manager: %v\n", err)
		os.Exit(1)
	}

	if err := mgr.CreateBaseSpec(); err != nil {
		fmt.Fprintf(os.Stderr, "create base spec: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote base CDI spec under %s\n", *outDir)

	if strings.TrimSpace(*indices) == "" {
		return
	}

	var devices []gcucdi.Device
	for _, part := range strings.Split(*indices, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		idx, err := strconv.Atoi(part)
		if err != nil {
			fmt.Fprintf(os.Stderr, "invalid index %q: %v\n", part, err)
			os.Exit(2)
		}
		devices = append(devices, gcucdi.Device{
			Name:       fmt.Sprintf("gcu-%d", idx),
			Index:      idx,
			VideoNodes: *videoNodes,
		})
	}
	if len(devices) == 0 {
		return
	}

	if err := mgr.CreateClaimSpec(*claimUID, devices); err != nil {
		fmt.Fprintf(os.Stderr, "create claim spec: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("wrote claim CDI spec for indices %s under %s\n", *indices, filepath.Clean(*outDir))
}
