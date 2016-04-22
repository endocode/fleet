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
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/coreos/fleet/functional/platform"
	"github.com/coreos/fleet/functional/util"
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

	if err := doMultipleUnitsCmd(cluster, m, "submit", 9); err != nil {
		t.Fatal(err)
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

	if err := doMultipleUnitsCmd(cluster, m, "load", 6); err != nil {
		t.Fatal(err)
	}
}

func TestUnitStart(t *testing.T) {
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

	if err := doMultipleUnitsCmd(cluster, m, "start", 3); err != nil {
		t.Fatal(err)
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
		unit := fmt.Sprintf("fixtures/units/hello@%02d.service", i)
		stdout, stderr, err := cluster.Fleetctl(m, "submit", unit)
		if err != nil {
			t.Fatalf("Failed to submit a batch of units: \nstdout: %s\nstder: %s\nerr: %v", stdout, stderr, err)
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
		n := path.Base(name)
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

func doMultipleUnitsCmd(cluster platform.Cluster, m platform.Member, cmd string, numUnits int) error {
	launchUnitsCmd := func(cmd string, numUnits int) (unitFiles []string, err error) {
		args := []string{cmd}
		for i := 0; i < numUnits; i++ {
			unitFile := fmt.Sprintf("fixtures/units/hello@%d.service", i+1)
			args = append(args, unitFile)
			unitFiles = append(unitFiles, path.Base(unitFile))
		}

		if stdout, stderr, err := cluster.Fleetctl(m, args...); err != nil {
			return nil,
				fmt.Errorf("Unable to %s batch of units: \nstdout: %s\nstderr: %s\nerr: %v",
					cmd, stdout, stderr, err)
		} else if strings.Contains(stderr, "Error") {
			return nil,
				fmt.Errorf("Failed to correctly %s batch of units: \nstdout: %s\nstderr: %s\nerr: %v",
					cmd, stdout, stderr, err)
		}

		return unitFiles, nil
	}

	checkListUnits := func(cmd string, unitFiles []string, inNumUnits int) (err error) {
		// wait until the unit gets processed up to 15 seconds
		if cmd == "submit" {
			listUnitStates, err := cluster.WaitForNUnitFiles(m, inNumUnits)
			if err != nil {
				return fmt.Errorf("Failed to run list-unit-files: %v", err)
			}

			if inNumUnits == 0 && len(listUnitStates) != 0 {
				return fmt.Errorf("Expected nil unit file list, got %v", listUnitStates)
			}

			// given unit name must be there in list-unit-files
			for i := 0; i < inNumUnits; i++ {
				_, found := listUnitStates[unitFiles[i]]
				if len(listUnitStates) != inNumUnits || !found {
					return fmt.Errorf("Expected %s to be unit file, got %v",
						unitFiles[i], listUnitStates)
				}
			}
		} else {
			// cmd == "load" or "start"
			var listUnitStates map[string][]util.UnitState
			if cmd == "load" {
				listUnitStates, err = cluster.WaitForNUnits(m, inNumUnits)
			} else {
				listUnitStates, err = cluster.WaitForNActiveUnits(m, inNumUnits)
			}
			if err != nil {
				return fmt.Errorf("Failed to run list-units: %v", err)
			}

			if inNumUnits == 0 && len(listUnitStates) != 0 {
				return fmt.Errorf("Expected nil unit list, got %v", listUnitStates)
			}

			// given unit name must be there in list-units
			for i := 0; i < inNumUnits; i++ {
				_, found := listUnitStates[unitFiles[i]]
				if len(listUnitStates) != inNumUnits || !found {
					return fmt.Errorf("Expected %s to be unit, got %v",
						unitFiles[i], listUnitStates)
				}
			}
		}

		return nil
	}

	cleanUnits := func(dcmd string, unitFile string) (err error) {
		if _, _, err := cluster.Fleetctl(m, dcmd, unitFile); err != nil {
			return fmt.Errorf("Failed to %s unit: %v", dcmd, err)
		}
		return nil
	}

	dcmd := make(map[string]string, 0)
	dcmd["submit"] = "destroy"
	dcmd["load"] = "unload"
	dcmd["start"] = "stop"

	// launch a batch of processing units
	unitFiles, err := launchUnitsCmd(cmd, numUnits)
	if err != nil {
		return err
	}
	if err := checkListUnits(cmd, unitFiles, numUnits); err != nil {
		return err
	}

	// destroy the unit and ensure it disappears from the unit list
	for i := 0; i < numUnits; i++ {
		if err := cleanUnits(dcmd[cmd], unitFiles[i]); err != nil {
			return err
		}
	}
	if err := checkListUnits(cmd, unitFiles, 0); err != nil {
		return err
	}

	// launch a batch of processing units
	unitFiles, err = launchUnitsCmd(cmd, numUnits)
	if err != nil {
		return err
	}
	if err := checkListUnits(cmd, unitFiles, numUnits); err != nil {
		return err
	}

	// destroy the unit again, not to affect the next tests for multiple units
	for i := 0; i < numUnits; i++ {
		if err := cleanUnits(dcmd[cmd], unitFiles[i]); err != nil {
			return err
		}
	}
	if err := checkListUnits(cmd, unitFiles, 0); err != nil {
		return err
	}

	return nil
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

	WaitForNUnitsCmd := func(cmd string, expectedUnits int) (err error) {
		if cmd == "submit" {
			_, err = cluster.WaitForNUnitFiles(m, expectedUnits)
		} else {
			_, err = cluster.WaitForNUnits(m, expectedUnits)
		}
		return err
	}

	// run a command for a unit and assert it shows up
	if _, _, err := cluster.Fleetctl(m, cmd, fxtHelloService); err != nil {
		return fmt.Errorf("Unable to %s fleet unit: %v", cmd, err)
	}
	if err := WaitForNUnitsCmd(cmd, 1); err != nil {
		return fmt.Errorf("Did not find 1 unit in cluster: %v", err)
	}

	helloFilename := path.Base(tmpHelloService)

	// store content of hello.service to bodyOrig
	bodyOrig, _, err := cluster.Fleetctl(m, "cat", helloFilename)
	if err != nil {
		return fmt.Errorf("Failed to run cat %s: %v", helloFilename, err)
	}

	// replace the unit and assert it shows up
	err = util.GenNewFleetService(tmpHelloService, fxtHelloService, "sleep 2", "sleep 1")
	if err != nil {
		return fmt.Errorf("Failed to generate a temp fleet service: %v", err)
	}
	if _, _, err := cluster.Fleetctl(m, cmd, "--replace", tmpHelloService); err != nil {
		return fmt.Errorf("Unable to replace fleet unit: %v", err)
	}
	if err := WaitForNUnitsCmd(cmd, 1); err != nil {
		return fmt.Errorf("Did not find 1 unit in cluster: %v", err)
	}

	// store content of the replaced unit hello.service to bodyNew
	bodyNew, _, err := cluster.Fleetctl(m, "cat", helloFilename)
	if err != nil {
		return fmt.Errorf("Failed to run cat %s: %v", helloFilename, err)
	}

	if bodyOrig == bodyNew {
		return fmt.Errorf("Error. the unit %s has not been replaced.", helloFilename)
	}

	os.Remove(tmpHelloService)

	if _, _, err := cluster.Fleetctl(m, "destroy", fxtHelloService); err != nil {
		return fmt.Errorf("Failed to destroy unit: %v", err)
	}
	if err := WaitForNUnitsCmd(cmd, 0); err != nil {
		return fmt.Errorf("Failed to get every unit to be cleaned up: %v", err)
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

	WaitForNUnitsCmd := func(cmd string, expectedUnits int) (err error) {
		if cmd == "submit" {
			_, err = cluster.WaitForNUnitFiles(m, expectedUnits)
		} else {
			_, err = cluster.WaitForNUnits(m, expectedUnits)
		}
		return err
	}

	if _, err := os.Stat(tmpFixtures); os.IsNotExist(err) {
		os.Mkdir(tmpFixtures, 0755)
	}

	var stdout string
	var bodiesOrig []string
	for i := 1; i <= n; i++ {
		curHelloService := fmt.Sprintf("/tmp/hello%d.service", i)
		tmpHelloFixture := fmt.Sprintf("/tmp/fixtures/hello%d.service", i)

		// generate a new service derived by fixtures, and store it under /tmp
		err = util.CopyFile(tmpHelloFixture, fxtHelloService)
		if err != nil {
			return fmt.Errorf("Failed to copy a temp fleet service: %v", err)
		}

		// run a command for a unit and assert it shows up
		if _, _, err := cluster.Fleetctl(m, cmd, tmpHelloFixture); err != nil {
			return fmt.Errorf("Unable to %s fleet unit: %v", cmd, err)
		}
		if err := WaitForNUnitsCmd(cmd, i); err != nil {
			return fmt.Errorf("Did not find %d units in cluster: \n%s", i, stdout)
		}

		helloFilename := path.Base(curHelloService)

		// retrieve content of hello.service, and append to bodiesOrig[]
		bodyCur, _, err := cluster.Fleetctl(m, "cat", helloFilename)
		if err != nil {
			return fmt.Errorf("Failed to run cat %s: %v", helloFilename, err)
		}

		bodiesOrig = append(bodiesOrig, bodyCur)

		// generate a new service derived by fixtures, and store it under /tmp
		err = util.GenNewFleetService(curHelloService, fxtHelloService, "sleep 2", "sleep 1")
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
		if err := WaitForNUnitsCmd(cmd, n); err != nil {
			return fmt.Errorf("Did not find %d units in cluster: \n%s", n, stdout)
		}

		helloFilename := path.Base(curHelloService)

		// retrieve content of hello.service, and compare it with the
		// correspondent entry in bodiesOrig[]
		bodyCur, _, err := cluster.Fleetctl(m, "cat", helloFilename)
		if err != nil {
			return fmt.Errorf("Failed to run cat %s: %v", helloFilename, err)
		}

		if bodiesOrig[i-1] == bodyCur {
			return fmt.Errorf("Error. the unit %s has not been replaced.", helloFilename)
		}
	}

	// clean up temp services under /tmp
	for i := 1; i <= n; i++ {
		curHelloService := fmt.Sprintf("/tmp/hello%d.service", i)

		if _, _, err := cluster.Fleetctl(m, "destroy", curHelloService); err != nil {
			fmt.Printf("Failed to destroy unit: %v", err)
			continue
		}

		os.Remove(curHelloService)
	}

	if err := WaitForNUnitsCmd(cmd, 0); err != nil {
		return fmt.Errorf("Failed to get every unit to be cleaned up: %v", err)
	}

	os.Remove(tmpFixtures)

	return nil
}
