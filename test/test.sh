#!/bin/bash

go generate ./...
git diff-index --quiet HEAD -- || (echo "Generation changed!" ; exit 1)
