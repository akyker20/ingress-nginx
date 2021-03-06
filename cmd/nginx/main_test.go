/*
Copyright 2017 The Kubernetes Authors.

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
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"k8s.io/ingress-nginx/internal/file"
	"k8s.io/ingress-nginx/internal/ingress/controller"
)

func TestCreateApiserverClient(t *testing.T) {
	home := os.Getenv("HOME")
	kubeConfigFile := fmt.Sprintf("%v/.kube/config", home)

	cli, err := createApiserverClient("", kubeConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error creating Kubernetes REST client: %v", err)
	}
	if cli == nil {
		t.Fatal("Expected a REST client but none returned.")
	}

	_, err = createApiserverClient("", "")
	if err == nil {
		t.Fatal("Expected an error creating REST client without an API server URL or kubeconfig file.")
	}
}

func TestHandleSigterm(t *testing.T) {
	home := os.Getenv("HOME")
	kubeConfigFile := fmt.Sprintf("%v/.kube/config", home)

	cli, err := createApiserverClient("", kubeConfigFile)
	if err != nil {
		t.Fatalf("Unexpected error creating Kubernetes REST client: %v", err)
	}

	resetForTesting(func() { t.Fatal("bad parse") })

	os.Setenv("POD_NAME", "test")
	os.Setenv("POD_NAMESPACE", "test")
	defer os.Setenv("POD_NAME", "")
	defer os.Setenv("POD_NAMESPACE", "")

	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--default-backend-service", "ingress-nginx/default-backend-http", "--http-port", "0", "--https-port", "0"}

	_, conf, err := parseFlags()
	if err != nil {
		t.Errorf("Unexpected error creating NGINX controller: %v", err)
	}
	conf.Client = cli

	fs, err := file.NewFakeFS()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	ngx := controller.NewNGINXController(conf, fs)

	go handleSigterm(ngx, func(code int) {
		if code != 1 {
			t.Errorf("Expected exit code 1 but %d received", code)
		}

		return
	})

	time.Sleep(1 * time.Second)

	t.Logf("Sending SIGTERM to PID %d", syscall.Getpid())
	err = syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	if err != nil {
		t.Error("Unexpected error sending SIGTERM signal.")
	}
}

func TestRegisterHandlers(t *testing.T) {
	// TODO
}
