SIGOZ_DIR ?= $(HOME)/Work/Space/Noizu/Projects/signoz

.PHONY: contract-check
contract-check:
	bash scripts/contract-check/run.sh "$(SIGOZ_DIR)"
