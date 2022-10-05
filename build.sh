#!/bin/bash

go build -ldflags "-X github.com/AssemblyAI/assemblyai-cli/cmd.PH_TOKEN=$POSTHOG_API_TOKEN" -o assemblyai