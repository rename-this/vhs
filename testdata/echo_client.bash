#!/bin/bash

go run echo.go &
sleep 1
go run echo_client.go
