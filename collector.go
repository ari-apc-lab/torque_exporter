/*
 * Copyright (C) 2019  Atos Spain SA. All rights reserved.
 *
 * This file is part of torque_exporter.
 *
 * torque_exporter is free software: you can redistribute it and/or modify it 
 * under the terms of the Apache License, Version 2.0 (the License);
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * The software is provided "AS IS", without any warranty of any kind, express 
 * or implied, including but not limited to the warranties of merchantability, 
 * fitness for a particular purpose and noninfringement, in no event shall the 
 * authors or copyright holders be liable for any claim, damages or other 
 * liability, whether in action of contract, tort or otherwise, arising from, 
 * out of or in connection with the software or the use or other dealings in the 
 * software.
 *
 * See DISCLAIMER file for the full disclaimer information and LICENSE and 
 * LICENSE-AGREEMENT files for full license information in the project root.
 *
 * Authors:  Atos Research and Innovation, Atos SPAIN SA
 */

package main

import (
	"bytes"
	"errors"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/spiros-atos/torque_exporter/ssh"

	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

const (
	sCOMPLETED	= iota
	sEXITING	= iota
	sHELD		= iota
	sQUEUED		= iota
	sRUNNING	= iota
	sMOVING		= iota
	sWAITING	= iota
	sSUSPENDED	= iota
)

/*
from man qstat:
 -  the job state:
       C -  Job is completed after having run/
       E -  Job is exiting after having run.
       H -  Job is held.
       Q -  job is queued, eligible to run or routed.
       R -  job is running.
       T -  job is being moved to new location.
       W -  job is waiting for its execution time
            (-a option) to be reached.
       S -  (Unicos only) job is suspend.
*/

// StatusDict maps string status with its int values
var StatusDict = map[string]int{
	"C":    sCOMPLETED,
	"E":	sEXITING,
	"H":	sHELD,
	"Q":	sQUEUED,
	"R":    sRUNNING,
	"T":    sMOVING,
	"W":    sWAITING,
	"S":    sSUSPENDED,
}

type TorqueCollector struct {
	queueRunning      *prometheus.Desc
	// queueCompleted    *prometheus.Desc
	userJobs          *prometheus.Desc
	// jobDetails        *prometheus.Desc
	partitionNodes    *prometheus.Desc
	sshConfig         *ssh.SSHConfig
	sshClient         *ssh.SSHClient
	timeZone          *time.Location
	alreadyRegistered []string
	lasttime          time.Time
}

func NewerTorqueCollector(host, sshUser, sshPass, 
		timeZone string) *TorqueCollector {
	newerTorqueCollector := &TorqueCollector{
		queueRunning: prometheus.NewDesc(
			"te_showq_r",
			"torque's queue",
			[]string{"jobid", "state", "username", "remaining", "starttime"},
			nil,
		),
		userJobs: prometheus.NewDesc(
			"te_qstat_u",
			"user's jobs",
			[]string{"jobid", "username", "jobname", "status"},
			nil,
		),
		partitionNodes: prometheus.NewDesc(
			"te_qstat_f",
			"job details",
			[]string{"partition", "availability", "state"},
			nil,
		),
		sshConfig: ssh.NewSSHConfigByPassword(
			sshUser,
			sshPass,
			host,
			22,
		),
		sshClient:         nil,
		alreadyRegistered: make([]string, 0),
	}
	var err error
	newerTorqueCollector.timeZone, err = time.LoadLocation(timeZone)
	if err != nil {
		log.Fatalln(err.Error())
	}
	newerTorqueCollector.setLastTime()
	return newerTorqueCollector
}

// Describe sends metrics descriptions of this collector
// through the ch channel.
// It implements collector interface
func (sc *TorqueCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- sc.queueRunning
	// ch <- sc.queueCompleted
	ch <- sc.userJobs
	// ch <- sc.jobDetails
	ch <- sc.partitionNodes
}

// Collect read the values of the metrics and
// passes them to the ch channel.
// It implements collector interface
func (sc *TorqueCollector) Collect(ch chan<- prometheus.Metric) {
	var err error
	sc.sshClient, err = sc.sshConfig.NewClient()
	if err != nil {
		log.Errorf("Creating SSH client: %s", err.Error())
		return
	}

	sc.collectQstat(ch)
	sc.collectQueue(ch)
	// sc.collectInfo(ch)

	err = sc.sshClient.Close()
	if err != nil {
		log.Errorf("Closing SSH client: %s", err.Error())
	}
}

func (sc *TorqueCollector) executeSSHCommand(cmd string) (*ssh.SSHSession, 
		error) {
	command := &ssh.SSHCommand{
		Path: cmd,
		// Env:    []string{"LC_DIR=/usr"},
	}

	var outb, errb bytes.Buffer
	session, err := sc.sshClient.OpenSession(nil, &outb, &errb)
	if err == nil {
		err = session.RunCommand(command)
		return session, err
	}
	return nil, err
}

func (sc *TorqueCollector) setLastTime() {
	sc.lasttime = time.Now().In(sc.timeZone).Add(-1 * time.Minute)
}

func parseTorqueTime(field string) (uint64, error) {
	var days, hours, minutes, seconds uint64
	var err error

	toParse := field
	haveDays := false

	// get days
	slice := strings.Split(toParse, "-")
	if len(slice) == 1 {
		toParse = slice[0]
	} else if len(slice) == 2 {
		days, err = strconv.ParseUint(slice[0], 10, 64)
		if err != nil {
			return 0, err
		}
		toParse = slice[1]
		haveDays = true
	} else {
		err = errors.New("Torque time could not be parsed: " + field)
		return 0, err
	}

	// get hours, minutes and seconds
	slice = strings.Split(toParse, ":")
	if len(slice) == 3 {
		hours, err = strconv.ParseUint(slice[0], 10, 64)
		if err == nil {
			minutes, err = strconv.ParseUint(slice[1], 10, 64)
			if err == nil {
				seconds, err = strconv.ParseUint(slice[1], 10, 64)
			}
		}
		if err != nil {
			return 0, err
		}
	} else if len(slice) == 2 {
		if haveDays {
			hours, err = strconv.ParseUint(slice[0], 10, 64)
			if err == nil {
				minutes, err = strconv.ParseUint(slice[1], 10, 64)
			}
		} else {
			minutes, err = strconv.ParseUint(slice[0], 10, 64)
			if err == nil {
				seconds, err = strconv.ParseUint(slice[1], 10, 64)
			}
		}
		if err != nil {
			return 0, err
		}
	} else if len(slice) == 1 {
		if haveDays {
			hours, err = strconv.ParseUint(slice[0], 10, 64)
		} else {
			minutes, err = strconv.ParseUint(slice[0], 10, 64)
		}
		if err != nil {
			return 0, err
		}
	} else {
		err = errors.New("Torque time could not be parsed: " + field)
		return 0, err
	}

	return days*24*60*60 + hours*60*60 + minutes*60 + seconds, nil
}

// nextLineIterator returns a function that iterates
// over an io.Reader object returning each line  parsed
// in fields following the parser method passed as argument
func nextLineIterator(buf io.Reader, 
		parser func(string) []string) func() ([]string, error) {
	var buffer = buf.(*bytes.Buffer)
	var parse = parser
	return func() ([]string, error) {
		// get next line in buffer
		line, err := buffer.ReadString('\n')
		if err != nil {
			return nil, err
		}
		// fmt.Print(line)

		// parse the line and return
		parsed := parse(line)
		if parsed == nil {
			return nil, errors.New("not able to parse line")
		}
		return parsed, nil
	}
}
