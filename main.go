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
	"flag"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	addr = flag.String(
		"listen-address",
		":9100",
		"The address to listen on for HTTP requests.",
	)
	host = flag.String(
		"host",
		"localhost",
		"Torque host torque domain name or IP.",
	)
	sshUser = flag.String(
		"ssh-user",
		"",
		"SSH user for remote torque connection (no localhost).",
	)
	sshPass = flag.String(
		"ssh-password",
		"",
		"SSH password for remote torque connection (no localhost).",
	)
	countryTZ = flag.String(
		"countrytz",
		"Europe/Madrid",
		"Country Time zone of the host, (e.g. \"Europe/Madrid\").",
	)
	logLevel = flag.String(
		"log-level",
		"error",
		"Log level of the Application.",
	)
)

func main() {
	flag.Parse()

	// Parse and set log lovel
	level, err := log.ParseLevel(*logLevel)
	if err == nil {
		log.SetLevel(level)
	} else {
		log.SetLevel(log.WarnLevel)
		log.Warnf("Log level %s not recognized, setting 'warn' as default.")
	}

	// Flags check
	if *host == "localhost" {
		flag.Usage()
		log.Fatalln("Localhost torque connection not implemented yet.")
	} else {
		if *sshUser == "" {
			flag.Usage()
			log.Fatalln(`A user must be provided to connect to Torque
				remotely.`)
		}
		if *sshPass == "" {
			flag.Usage()
			log.Warnln(`A password should be provided to connect to Torque
				remotely.`)
		}
	}

	prometheus.MustRegister(NewerTorqueCollector(*host, *sshUser, *sshPass, 
		*countryTZ))

	// Expose the registered metrics via HTTP.
	log.Infof("Starting Server: %s", *addr)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(*addr, nil))
}
