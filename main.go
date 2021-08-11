package main

// ------------------------------------------------------------------------------
// MIT License

// Copyright (c) 2020 deepfence

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
// ------------------------------------------------------------------------------

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/lkarlslund/SecretScanner/core"
	"github.com/lkarlslund/SecretScanner/output"
	"github.com/lkarlslund/SecretScanner/signature"
)

// Read the regex signatures from config file, options etc.
// and setup the session to start scanning for secrets
var session = core.GetSession()

// Scan a container image for secrets layer by layer
// @parameters
// image - Name of the container image to scan (e.g. "alpine:3.5")
// @returns
// Error, if any. Otherwise, returns nil
func findSecretsInImage(image string) error {
	var tempSecretsFound []output.SecretFound

	tempDir, err := core.GetTmpDir(image) // ("Deepfence/SecretScanning/" + scanId)
	if err != nil {
		core.GetSession().Log.Error("findSecretsInImage: Could not create temp dir: %s", err)
		return err
	}
	defer core.DeleteTmpDir(tempDir)

	imageScan := ImageScan{imageName: image, imageId: "", tempDir: tempDir}
	err = imageScan.extractImage()

	if err != nil {
		core.GetSession().Log.Error("findSecretsInImage: %s", err)
		return err
	}

	jsonImageSecretsOutput := output.JsonImageSecretsOutput{ImageName: image}
	jsonImageSecretsOutput.SetTime()
	jsonImageSecretsOutput.SetImageId(imageScan.imageId)
	jsonImageSecretsOutput.PrintJsonHeader()

	tempSecretsFound, err = imageScan.scan()

	if err != nil {
		core.GetSession().Log.Error("findSecretsInImage: %s", err)
		return err
	}

	jsonImageSecretsOutput.PrintJsonFooter()

	jsonImageSecretsOutput.SetSecrets(tempSecretsFound)
	jsonFilename, err := core.GetJsonFilepath(image)
	if err != nil {
		core.GetSession().Log.Error("findSecretsInImage: %s", err)
		return err
	}
	err = jsonImageSecretsOutput.WriteSecrets(jsonFilename)
	if err != nil {
		core.GetSession().Log.Error("findSecretsInImage: %s", err)
		return err
	}

	return nil
}

// Scan a container image for secrets layer by layer
// @parameters
// dir - Complete path of the directory to be scanned
// @returns
// Error, if any. Otherwise, returns nil
func findSecretsInDir(dir string) error {
	var isFirstSecret bool = true
	var numSecrets uint = 0

	jsonDirSecretsOutput := output.JsonDirSecretsOutput{DirName: *session.Options.Local}
	jsonDirSecretsOutput.SetTime()
	jsonDirSecretsOutput.PrintJsonHeader()

	secrets, err := scanSecretsInDir("", "", dir, &isFirstSecret, &numSecrets)
	if err != nil {
		core.GetSession().Log.Error("findSecretsInDir: %s", err)
		return err
	}

	jsonDirSecretsOutput.PrintJsonFooter()

	jsonDirSecretsOutput.SetSecrets(secrets)
	jsonFilename, err := core.GetJsonFilepath(dir)
	if err != nil {
		core.GetSession().Log.Error("findSecretsInDir: %s", err)
		return err
	}
	err = jsonDirSecretsOutput.WriteSecrets(jsonFilename)
	if err != nil {
		core.GetSession().Log.Error("findSecretsInDir: %s", err)
		return err
	}

	return nil
}

func main() {
	// Process and store the read signatures
	signature.ProcessSignatures(session.Config.Signatures)

	// Build Hyperscan database for fast scanning
	signature.BuildHsDb()

	// Scan container image for secrets
	if len(*session.Options.ImageName) > 0 {
		fmt.Printf("Scanning image %s for secrets...\n", *session.Options.ImageName)
		err := findSecretsInImage(*session.Options.ImageName)
		if err != nil {
			core.GetSession().Log.Fatal("main: error while scanning image: %s", err)
		}
	}

	// Scan local directory for secrets
	if len(*session.Options.Local) > 0 {
		fmt.Printf("[*] Scanning local directory: %s\n", color.BlueString(*session.Options.Local))
		err := findSecretsInDir(*session.Options.Local)
		if err != nil {
			core.GetSession().Log.Fatal("main: error while scanning dir: %s", err)
		}
	}
}
