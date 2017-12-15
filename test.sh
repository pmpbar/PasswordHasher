#!/bin/bash
go test -run Hash && \
go test -run Multi && \
go test -run Shutdown
