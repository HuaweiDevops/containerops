/*
Copyright 2014 - 2017 Huawei Technologies Co., Ltd. All rights reserved.

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
	"os/exec"
	"path"
	"strings"

	"gopkg.in/src-d/go-git.v4"
)

const (
	FAILUER_EXIT      = -1
	MISSING_PARAMATER = -2
	PARSE_ENV_FAILURE = -3
	CLONE_ERROR       = -4
	UNKNOWN_ACTION    = -5
)

//Parse CO_DATA value, and return Kubernetes repository URI and action (build/test/publish).
func parse_env(env string) (uri string, action string, err error) {
	files := strings.Fields(env)
	if len(files) == 0 {
		return "", "", fmt.Errorf("CO_DATA value is null")
	}

	for _, v := range files {
		s := strings.Split(v, "=")
		key, value := s[0], s[1]

		switch key {
		case "kubernetes":
			uri = value
		case "action":
			action = value
		default:
			fmt.Fprintf(os.Stdout, "[COUT] Unknown Parameter: [%s]", s)
		}
	}

	return uri, action, nil
}

//Git clone the kubernetes repository, and process will redirect to system stdout.
func git_clone(repo, dest string) error {
	if _, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:               repo,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		Progress:          os.Stdout,
	}); err != nil {
		return err
	}

	return nil
}

//`make bazel-test`
func bazel_test(dest string) {
	cmd := exec.Command("make bazel-test")
	cmd.Path = dest
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[COUT] Bazel test error: %s", err.Error())
		fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = false")
		os.Exit(FAILUER_EXIT)
	}
}

//`make bazel-build`
func bazel_build(dest string) {
	cmd := exec.Command("make bazel-build")
	cmd.Path = dest
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "[COUT] Bazel build error: %s", err.Error())
		fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = false")
		os.Exit(FAILUER_EXIT)
	}
}

//TODO Build the kubernetes all binrary files, and publish to containerops repository. And not execute the `make bazel-publish` command.
func publish(dest string) {

}

func main() {
	//Get the CO_DATA from environment parameter "CO_DATA"
	co_data := os.Getenv("CO_DATA")
	if len(co_data) == 0 {
		fmt.Fprintf(os.Stderr, "[COUT] The CO_DATA value is null.")
		fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = false")
		os.Exit(MISSING_PARAMATER)
	}

	//Parse the CO_DATA, get the kubernetes repository URI and action
	if k8s_repo, action, err := parse_env(co_data); err != nil {
		fmt.Fprintf(os.Stderr, "[COUT] Parse the CO_DATA error: %s", err.Error())
		fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = false")
		os.Exit(PARSE_ENV_FAILURE)
	} else {
		//Create the base path within GOPATH.
		base_path := path.Join(os.Getenv("GOPATH"), "github.com", "kubernetes")
		os.MkdirAll(base_path, 0777)

		//Clone the git repository
		if err := git_clone(k8s_repo, base_path); err != nil {
			fmt.Fprintf(os.Stderr, "[COUT] Clone the kubernetes repository error: %s", err.Error())
			fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = false")
			os.Exit(CLONE_ERROR)
		}

		k8s_path := path.Join(base_path, "kubernetes")

		//Execute action
		switch action {
		case "build":
			bazel_build(k8s_path)
		case "test":
			bazel_test(k8s_path)
		case "publish":
			publish(k8s_path)
		default:
			fmt.Fprintf(os.Stderr, "[COUT] Unknown action, the component only support build, test and publish action.")
			fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = false")
			os.Exit(UNKNOWN_ACTION)
		}

	}

	//Print result
	fmt.Fprintf(os.Stdout, "[COUT] CO_RESULT = true")
	os.Exit(0)
}
