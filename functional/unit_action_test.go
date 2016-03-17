// Copyright 2014 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package functional

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/coreos/fleet/functional/platform"
)

const (
	tmpHelloService = "/tmp/hello.service"
	fxtHelloService = "fixtures/units/hello.service"
	tmpFixtures     = "/tmp/fixtures"
	numUnitsReplace = 9
)

// TestUnitRunnable is the simplest test possible, deplying a single-node
// cluster and ensuring a unit can enter an 'active' state
func TestUnitRunnable(t *testing.T) {
	cluster, err := platform.NewNspawnCluster("smoke")
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Destroy()

	m0, err := cluster.CreateMember()
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.WaitForNMachines(m0, 1)
	if err != nil {
		t.Fatal(err)
	}

	if stdout, stderr, err := cluster.Fleetctl(m0, "start", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Unable to start fleet unit: \nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}

	units, err := cluster.WaitForNActiveUnits(m0, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, found := units["hello.service"]
	if len(units) != 1 || !found {
		t.Fatalf("Expected hello.service to be sole active unit, got %v", units)
	}
}

func TestUnitSubmit(t *testing.T) {
	cluster, err := platform.NewNspawnCluster("smoke")
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Destroy()

	m, err := cluster.CreateMember()
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.WaitForNMachines(m, 1)
	if err != nil {
		t.Fatal(err)
	}

	// submit a unit and assert it shows up
	if _, _, err := cluster.Fleetctl(m, "submit", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Unable to submit fleet unit: %v", err)
	}
	stdout, _, err := cluster.Fleetctl(m, "list-units", "--no-legend")
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}
	units := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(units) != 1 {
		t.Fatalf("Did not find 1 unit in cluster: \n%s", stdout)
	}

	// submitting the same unit should not fail
	if _, _, err = cluster.Fleetctl(m, "submit", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Expected no failure when double-submitting unit, got this: %v", err)
	}

	// destroy the unit and ensure it disappears from the unit list
	if _, _, err := cluster.Fleetctl(m, "destroy", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Failed to destroy unit: %v", err)
	}
	stdout, _, err = cluster.Fleetctl(m, "list-units", "--no-legend")
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}
	if strings.TrimSpace(stdout) != "" {
		t.Fatalf("Did not find 0 units in cluster: \n%s", stdout)
	}

	// submitting the unit after destruction should succeed
	if _, _, err := cluster.Fleetctl(m, "submit", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Unable to submit fleet unit: %v", err)
	}
	stdout, _, err = cluster.Fleetctl(m, "list-units", "--no-legend")
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}
	units = strings.Split(strings.TrimSpace(stdout), "\n")
	if len(units) != 1 {
		t.Fatalf("Did not find 1 unit in cluster: \n%s", stdout)
	}
}

func TestUnitRestart(t *testing.T) {
	cluster, err := platform.NewNspawnCluster("smoke")
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Destroy()

	m, err := cluster.CreateMember()
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.WaitForNMachines(m, 1)
	if err != nil {
		t.Fatal(err)
	}

	if stdout, stderr, err := cluster.Fleetctl(m, "start", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Unable to start fleet unit: \nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}

	units, err := cluster.WaitForNActiveUnits(m, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, found := units["hello.service"]
	if len(units) != 1 || !found {
		t.Fatalf("Expected hello.service to be sole active unit, got %v", units)
	}

	if _, _, err := cluster.Fleetctl(m, "stop", "hello.service"); err != nil {
		t.Fatal(err)
	}
	units, err = cluster.WaitForNActiveUnits(m, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 0 {
		t.Fatalf("Zero units should be running, found %v", units)
	}

	if stdout, stderr, err := cluster.Fleetctl(m, "start", "hello.service"); err != nil {
		t.Fatalf("Unable to start fleet unit: \nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}
	units, err = cluster.WaitForNActiveUnits(m, 1)
	if err != nil {
		t.Fatal(err)
	}
	_, found = units["hello.service"]
	if len(units) != 1 || !found {
		t.Fatalf("Expected hello.service to be sole active unit, got %v", units)
	}

}

// TestUnitSubmitReplace() tests whether a command "fleetctl submit --replace
// hello.service" works or not.
func TestUnitSubmitReplace(t *testing.T) {
	if err := replaceUnitCommon("submit"); err != nil {
		t.Fatal(err)
	}

	if err := replaceUnitMultiple("submit", numUnitsReplace); err != nil {
		t.Fatal(err)
	}
}

// TestUnitLoadReplace() tests whether a command "fleetctl load --replace
// hello.service" works or not.
func TestUnitLoadReplace(t *testing.T) {
	if err := replaceUnitCommon("load"); err != nil {
		t.Fatal(err)
	}

	if err := replaceUnitMultiple("load", numUnitsReplace); err != nil {
		t.Fatal(err)
	}
}

// TestUnitStartReplace() tests whether a command "fleetctl start --replace
// hello.service" works or not.
func TestUnitStartReplace(t *testing.T) {
	if err := replaceUnitCommon("start"); err != nil {
		t.Fatal(err)
	}

	if err := replaceUnitMultiple("start", numUnitsReplace); err != nil {
		t.Fatal(err)
	}
}

func TestUnitSSHActions(t *testing.T) {
	cluster, err := platform.NewNspawnCluster("smoke")
	if err != nil {
		t.Fatal(err)
	}
	defer cluster.Destroy()

	m, err := cluster.CreateMember()
	if err != nil {
		t.Fatal(err)
	}
	_, err = cluster.WaitForNMachines(m, 1)
	if err != nil {
		t.Fatal(err)
	}

	if stdout, stderr, err := cluster.Fleetctl(m, "start", "--no-block", "fixtures/units/hello.service"); err != nil {
		t.Fatalf("Unable to start fleet unit: \nstdout: %s\nstderr: %s\nerr: %v", stdout, stderr, err)
	}

	units, err := cluster.WaitForNActiveUnits(m, 1)
	if err != nil {
		t.Fatal(err)
	}

	_, found := units["hello.service"]
	if len(units) != 1 || !found {
		t.Fatalf("Expected hello.service to be sole active unit, got %v", units)
	}

	stdout, _, err := cluster.Fleetctl(m, "--strict-host-key-checking=false", "ssh", "hello.service", "echo", "foo")
	if err != nil {
		t.Errorf("Failure occurred while calling fleetctl ssh: %v", err)
	}

	if !strings.Contains(stdout, "foo") {
		t.Errorf("Could not find expected string in command output:\n%s", stdout)
	}

	stdout, _, err = cluster.Fleetctl(m, "--strict-host-key-checking=false", "status", "hello.service")
	if err != nil {
		t.Errorf("Failure occurred while calling fleetctl status: %v", err)
	}

	if !strings.Contains(stdout, "Active: active") {
		t.Errorf("Could not find expected string in status output:\n%s", stdout)
	}

	stdout, _, err = cluster.Fleetctl(m, "--strict-host-key-checking=false", "journal", "--sudo", "hello.service")
	if err != nil {
		t.Errorf("Failure occurred while calling fleetctl journal: %v", err)
	}

	if !strings.Contains(stdout, "Hello, World!") {
		t.Errorf("Could not find expected string in journal output:\n%s", stdout)
	}
}

// replaceUnitCommon() tests whether a command "fleetctl {submit,load,start}
// --replace hello.service" works or not.
func replaceUnitCommon(cmd string) error {
	// check if cmd is one of the supported commands.
	listCmds := []string{"submit", "load", "start"}
	found := false
	for _, ccmd := range listCmds {
		if ccmd == cmd {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("invalid command %s", cmd)
	}

	cluster, err := platform.NewNspawnCluster("smoke")
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	defer cluster.Destroy()

	m, err := cluster.CreateMember()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	_, err = cluster.WaitForNMachines(m, 1)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	// run a command for a unit and assert it shows up
	if _, _, err := cluster.Fleetctl(m, cmd, fxtHelloService); err != nil {
		return fmt.Errorf("Unable to %s fleet unit: %v", cmd, err)
	}
	stdout, _, err := cluster.Fleetctl(m, "list-units", "--no-legend")
	if err != nil {
		return fmt.Errorf("Failed to run list-units: %v", err)
	}
	units := strings.Split(strings.TrimSpace(stdout), "\n")
	if len(units) != 1 {
		return fmt.Errorf("Did not find 1 unit in cluster: \n%s", stdout)
	}

	// replace the unit and assert it shows up
	err = genNewFleetService(tmpHelloService, fxtHelloService, "sleep 2", "sleep 1")
	if err != nil {
		return fmt.Errorf("Failed to generate a temp fleet service: %v", err)
	}
	if _, _, err := cluster.Fleetctl(m, cmd, "--replace", tmpHelloService); err != nil {
		return fmt.Errorf("Unable to replace fleet unit: %v", err)
	}
	stdout, _, err = cluster.Fleetctl(m, "list-units", "--no-legend")
	if err != nil {
		return fmt.Errorf("Failed to run list-units: %v", err)
	}
	units = strings.Split(strings.TrimSpace(stdout), "\n")
	if len(units) != 1 {
		return fmt.Errorf("Did not find 1 unit in cluster: \n%s", stdout)
	}
	os.Remove(tmpHelloService)

	if err := destroyUnitRetrying(cluster, m, fxtHelloService); err != nil {
		return fmt.Errorf("Cannot destroy unit %v", fxtHelloService)
	}

	return nil
}

// replaceUnitMultiple() tests whether a command "fleetctl {submit,load,start}
// --replace hello.service" works or not.
func replaceUnitMultiple(cmd string, n int) error {
	// check if cmd is one of the supported commands.
	listCmds := []string{"submit", "load", "start"}
	found := false
	for _, ccmd := range listCmds {
		if ccmd == cmd {
			found = true
		}
	}
	if !found {
		return fmt.Errorf("invalid command %s", cmd)
	}

	cluster, err := platform.NewNspawnCluster("smoke")
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	defer cluster.Destroy()

	m, err := cluster.CreateMember()
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	_, err = cluster.WaitForNMachines(m, 1)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if _, err := os.Stat(tmpFixtures); os.IsNotExist(err) {
		os.Mkdir(tmpFixtures, 0755)
	}

	var stdout string
	for i := 1; i <= n; i++ {
		curHelloService := fmt.Sprintf("/tmp/hello%d.service", i)
		tmpHelloFixture := fmt.Sprintf("/tmp/fixtures/hello%d.service", i)

		// generate a new service derived by fixtures, and store it under /tmp
		err = copyFile(tmpHelloFixture, fxtHelloService)
		if err != nil {
			return fmt.Errorf("Failed to copy a temp fleet service: %v", err)
		}

		// run a command for a unit and assert it shows up
		if _, _, err := cluster.Fleetctl(m, cmd, tmpHelloFixture); err != nil {
			return fmt.Errorf("Unable to %s fleet unit: %v", cmd, err)
		}

		stdout, _, err = cluster.Fleetctl(m, "list-unit-files", "--no-legend")
		if err != nil {
			return fmt.Errorf("Failed to run %s: %v", "list-unit-files", err)
		}
		units := strings.Split(strings.TrimSpace(stdout), "\n")
		if len(units) != i {
			return fmt.Errorf("Did not find %d units in cluster: \n%s", i, stdout)
		}

		// generate a new service derived by fixtures, and store it under /tmp
		err = genNewFleetService(curHelloService, fxtHelloService, "sleep 2", "sleep 1")
		if err != nil {
			return fmt.Errorf("Failed to generate a temp fleet service: %v", err)
		}
	}

	for i := 1; i <= n; i++ {
		curHelloService := fmt.Sprintf("/tmp/hello%d.service", i)

		// replace the unit and assert it shows up
		if _, _, err = cluster.Fleetctl(m, cmd, "--replace", curHelloService); err != nil {
			return fmt.Errorf("Unable to replace fleet unit: %v", err)
		}
		stdout, _, err = cluster.Fleetctl(m, "list-unit-files", "--no-legend")
		if err != nil {
			return fmt.Errorf("Failed to run %s: %v", "list-unit-files", err)
		}
		units := strings.Split(strings.TrimSpace(stdout), "\n")
		if len(units) != n {
			return fmt.Errorf("Did not find %d units in cluster: \n%s", n, stdout)
		}
	}

	// clean up temp services under /tmp
	for i := 1; i <= n; i++ {
		curHelloService := fmt.Sprintf("/tmp/hello%d.service", i)
		os.Remove(curHelloService)

		if err := destroyUnitRetrying(cluster, m, fxtHelloService); err != nil {
			return fmt.Errorf("Cannot destroy unit %v", fxtHelloService)
		}
	}
	os.Remove(tmpFixtures)

	return nil
}

// copyFile()
func copyFile(newFile, oldFile string) error {
	input, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(newFile, []byte(input), 0644)
	if err != nil {
		return err
	}
	return nil
}

// genNewFleetService() is a helper for generating a temporary fleet service
// that reads from oldFile, replaces oldVal with newVal, and stores the result
// to newFile.
func genNewFleetService(newFile, oldFile, newVal, oldVal string) error {
	input, err := ioutil.ReadFile(oldFile)
	if err != nil {
		return err
	}
	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, oldVal) {
			lines[i] = strings.Replace(line, oldVal, newVal, len(oldVal))
		}
	}
	output := strings.Join(lines, "\n")
	err = ioutil.WriteFile(newFile, []byte(output), 0644)
	if err != nil {
		return err
	}
	return nil
}

// destroyUnitRetrying() destroys the unit and ensure it disappears from the
// unit list. It could take a little time until the unit gets destroyed.
func destroyUnitRetrying(cluster platform.Cluster, m platform.Member, serviceFile string) error {
	maxAttempts := 3
	found := false
	var stdout string
	var err error
	for {
		if _, _, err := cluster.Fleetctl(m, "destroy", serviceFile); err != nil {
			return fmt.Errorf("Failed to destroy unit: %v", err)
		}
		stdout, _, err = cluster.Fleetctl(m, "list-units", "--no-legend")
		if err != nil {
			return fmt.Errorf("Failed to run list-units: %v", err)
		}
		if strings.TrimSpace(stdout) == "" || maxAttempts == 0 {
			found = true
			break
		}
		maxAttempts--
	}

	if !found {
		return fmt.Errorf("Did not find 0 units in cluster: \n%s", stdout)
	}

	return nil
}
