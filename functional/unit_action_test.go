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
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/coreos/fleet/functional/platform"
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

// TestUnitSubmit checks if a unit becomes submitted and destroyed successfully.
// First it submits a unit, and destroys the unit, verifies it's destroyed,
// finally submits the unit again.
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

	unitFile := "fixtures/units/hello.service"

	// submit a unit and assert it shows up
	if _, _, err := cluster.Fleetctl(m, "submit", unitFile); err != nil {
		t.Fatalf("Unable to submit fleet unit: %v", err)
	}

	// wait until the unit gets submitted up to 15 seconds
	listUnitStates, err := cluster.WaitForNUnitFiles(m, 1)
	if err != nil {
		t.Fatalf("Failed to run list-unit-files: %v", err)
	}

	// given unit name must be there in list-unit-files
	_, found := listUnitStates[path.Base(unitFile)]
	if len(listUnitStates) != 1 || !found {
		t.Fatalf("Expected %s to be unit file, got %v", path.Base(unitFile), listUnitStates)
	}

	// submitting the same unit should not fail
	if _, _, err = cluster.Fleetctl(m, "submit", unitFile); err != nil {
		t.Fatalf("Expected no failure when double-submitting unit, got this: %v", err)
	}

	// destroy the unit and ensure it disappears from the unit list
	if _, _, err := cluster.Fleetctl(m, "destroy", unitFile); err != nil {
		t.Fatalf("Failed to destroy unit: %v", err)
	}
	// wait until the unit gets destroyed up to 15 seconds
	listUnitStates, err = cluster.WaitForNUnitFiles(m, 0)
	if err != nil {
		t.Fatalf("Failed to run list-unit-files: %v", err)
	}
	if len(listUnitStates) != 0 {
		t.Fatalf("Expected nil unit file list, got %v", listUnitStates)
	}

	// submitting the unit after destruction should succeed
	if _, _, err := cluster.Fleetctl(m, "submit", unitFile); err != nil {
		t.Fatalf("Unable to submit fleet unit: %v", err)
	}

	// wait until the unit gets submitted up to 15 seconds
	listUnitStates, err = cluster.WaitForNUnitFiles(m, 1)
	if err != nil {
		t.Fatalf("Failed to run list-unit-files: %v", err)
	}

	// given unit name must be there in list-unit-files
	_, found = listUnitStates[path.Base(unitFile)]
	if len(listUnitStates) != 1 || !found {
		t.Fatalf("Expected %s to be unit file, got %v", path.Base(unitFile), listUnitStates)
	}
}

// TestUnitLoad checks if a unit becomes loaded and unloaded successfully.
// First it load a unit, and unloads the unit, verifies it's unloaded,
// finally loads the unit again.
func TestUnitLoad(t *testing.T) {
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

	unitFile := "fixtures/units/hello.service"

	// load a unit and assert it shows up
	_, _, err = cluster.Fleetctl(m, "load", unitFile)
	if err != nil {
		t.Fatalf("Unable to load fleet unit: %v", err)
	}

	// wait until the unit gets loaded up to 15 seconds
	listUnitStates, err := cluster.WaitForNUnits(m, 1)
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}

	// given unit name must be there in list-units
	_, found := listUnitStates[path.Base(unitFile)]
	if len(listUnitStates) != 1 || !found {
		t.Fatalf("Expected %s to be unit, got %v", path.Base(unitFile), listUnitStates)
	}

	// unload the unit and ensure it disappears from the unit list
	_, _, err = cluster.Fleetctl(m, "unload", unitFile)
	if err != nil {
		t.Fatalf("Failed to unload unit: %v", err)
	}

	// wait until the unit gets unloaded up to 15 seconds
	listUnitStates, err = cluster.WaitForNUnits(m, 0)
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}

	// given unit name must be there in list-units
	if len(listUnitStates) != 0 {
		t.Fatalf("Expected nil unit list, got %v", listUnitStates)
	}

	// loading the unit after destruction should succeed
	_, _, err = cluster.Fleetctl(m, "load", unitFile)
	if err != nil {
		t.Fatalf("Unable to load fleet unit: %v", err)
	}

	// wait until the unit gets loaded up to 15 seconds
	listUnitStates, err = cluster.WaitForNUnits(m, 1)
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}

	// given unit name must be there in list-units
	_, found = listUnitStates[path.Base(unitFile)]
	if len(listUnitStates) != 1 || !found {
		t.Fatalf("Expected %s to be unit, got %v", path.Base(unitFile), listUnitStates)
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

	stdout, stderr, err := cluster.Fleetctl(m, "--strict-host-key-checking=false", "ssh", "hello.service", "echo", "foo")
	if err != nil {
		t.Errorf("Failure occurred while calling fleetctl ssh: %v\nstdout: %v\nstderr: %v", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "foo") {
		t.Errorf("Could not find expected string in command output:\n%s", stdout)
	}

	stdout, stderr, err = cluster.Fleetctl(m, "--strict-host-key-checking=false", "status", "hello.service")
	if err != nil {
		t.Errorf("Failure occurred while calling fleetctl status: %v\nstdout: %v\nstderr: %v", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "Active: active") {
		t.Errorf("Could not find expected string in status output:\n%s", stdout)
	}

	stdout, stderr, err = cluster.Fleetctl(m, "--strict-host-key-checking=false", "journal", "--sudo", "hello.service")
	if err != nil {
		t.Errorf("Failure occurred while calling fleetctl journal: %v\nstdout: %v\nstderr: %v", err, stdout, stderr)
	}

	if !strings.Contains(stdout, "Hello, World!") {
		t.Errorf("Could not find expected string in journal output:\n%s", stdout)
	}
}

// TestUnitCat simply compares body of a unit file with that of a unit fetched
// from the remote cluster using "fleetctl cat".
func TestUnitCat(t *testing.T) {
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

	// read a sample unit file to a buffer
	unitFile := "fixtures/units/hello.service"
	fileBuf, err := ioutil.ReadFile(unitFile)
	if err != nil {
		t.Fatal(err)
	}
	fileBody := strings.TrimSpace(string(fileBuf))

	// submit a unit and assert it shows up
	_, _, err = cluster.Fleetctl(m, "submit", unitFile)
	if err != nil {
		t.Fatalf("Unable to submit fleet unit: %v", err)
	}
	// wait until the unit gets submitted up to 15 seconds
	_, err = cluster.WaitForNUnitFiles(m, 1)
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}

	// cat the unit file and compare it with the original unit body
	stdout, _, err := cluster.Fleetctl(m, "cat", path.Base(unitFile))
	if err != nil {
		t.Fatalf("Unable to submit fleet unit: %v", err)
	}
	catBody := strings.TrimSpace(stdout)

	if strings.Compare(catBody, fileBody) != 0 {
		t.Fatalf("unit body changed across fleetctl cat: \noriginal:%s\nnew:%s", fileBody, catBody)
	}
}

// TestUnitStatus simply checks "fleetctl status hello.service" actually works.
func TestUnitStatus(t *testing.T) {
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

	unitFile := "fixtures/units/hello.service"

	// Load a unit and print out status.
	// Without loading a unit, it's impossible to run fleetctl status
	_, _, err = cluster.Fleetctl(m, "load", unitFile)
	if err != nil {
		t.Fatalf("Unable to load a fleet unit: %v", err)
	}

	// wait until the unit gets loaded up to 15 seconds
	_, err = cluster.WaitForNUnits(m, 1)
	if err != nil {
		t.Fatalf("Failed to run list-units: %v", err)
	}

	stdout, stderr, err := cluster.Fleetctl(m,
		"--strict-host-key-checking=false", "status", path.Base(unitFile))
	if !strings.Contains(stdout, "Loaded: loaded") {
		t.Errorf("Could not find expected string in status output:\n%s\nstderr:\n%s",
			stdout, stderr)
	}
}

// TestListUnitFilesOrder simply checks if "fleetctl list-unit-files" returns
// an ordered list of units
func TestListUnitFilesOrder(t *testing.T) {
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

	// Combine units
	var units []string
	for i := 1; i <= 20; i++ {
		unit := fmt.Sprintf("fixtures/units/hello@%d.service", i)
		stdout, stderr, err := cluster.Fleetctl(m, "load", unit)
		if err != nil {
			t.Fatalf("Failed to load a batch of units: \nstdout: %s\nstder: %s\nerr: %v", stdout, stderr, err)
		}
		units = append(units, unit)
	}

	// make sure that all unit files will show up
	_, err = cluster.WaitForNUnitFiles(m, 20)
	if err != nil {
		t.Fatal("Failed to run list-unit-files: %v", err)
	}

	stdout, _, err := cluster.Fleetctl(m, "list-unit-files", "--no-legend", "--fields", "unit")
	if err != nil {
		t.Fatal("Failed to run list-unit-files: %v", err)
	}

	outUnits := strings.Split(strings.TrimSpace(stdout), "\n")

	var sortable sort.StringSlice
	for _, name := range units {
		n := strings.Replace(name, "fixtures/units/", "", 1)
		sortable = append(sortable, n)
	}
	sortable.Sort()

	var inUnits []string
	for _, name := range sortable {
		inUnits = append(inUnits, name)
	}

	if !reflect.DeepEqual(inUnits, outUnits) {
		t.Fatalf("Failed to get a sorted list of units from list-unit-files")
	}
}
