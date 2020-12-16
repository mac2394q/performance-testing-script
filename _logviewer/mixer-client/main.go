/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"context"
	"io"
	"time"

	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	log "k8s.io/klog"
	mixerPb "k8s.io/perf-tests/logviewer/mixer/request"
)

const (
	address        = "localhost:17655"
	timeoutSeconds = 300
)

func main() {
	log.InitFlags(nil)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := mixerPb.NewMixerServiceClient(conn)
	var req = getSampleRequest()
	response, err := client.DoWork(context.Background(), req)

	batchCount := 0
	for {
		workResult, err := response.Recv()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Errorf("Error receiving result %v", err)
		}
		if workResult.LogLines == nil {
			return
		}
		for _, line := range workResult.LogLines {
			if line == nil {
				log.Infof("Nil batch %v", batchCount)
			}
			log.Infof("%v : %v", line.Timestamp, line.Entry)
		}
		batchCount++
	}
}

func getSampleRequest() *mixerPb.MixerRequest {
	since, _ := time.Parse(time.RFC3339Nano, "2019-02-15T15:38:48.908485Z")
	until, _ := time.Parse(time.RFC3339Nano, "2019-02-15T18:38:48.908485Z")
	pSince, _ := ptypes.TimestampProto(since)
	pUntil, _ := ptypes.TimestampProto(until)

	return &mixerPb.MixerRequest{
		BuildNumber:     310,
		FilePrefix:      "kube-apiserver-audit.log-",
		TargetSubstring: "9a27",
		Since:           pSince,
		Until:           pUntil,
	}
}
