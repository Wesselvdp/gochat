
# Load .env file if it exists
ifneq (,$(wildcard .env))
    include scripts/.env
    export
endif

#ENV ?= development  # Default to development
PRODUCTION_URL = https://app.torgon.io
STAGING_URL = https://staging.torgon.io
DEV_URL = http://localhost:8080

DOKPLOY_APP := $(if $(filter $(ENV),production),newproject-gochat-f51808,gochat-app-staging-bd0bf7)
DB_PATH = /etc/dokploy/compose/$(DOKPLOY_APP)/files/sqlite/database.db
# Determine URL based on the environment
URL := $(if $(filter $(ENV),production),$(PRODUCTION_URL),$(if $(filter $(ENV),staging),$(STAGING_URL),$(if $(filter $(ENV),development),$(DEV_URL), "") ))
DOCKER_CONTAINER := $(if $(filter $(ENV),production),adfed9fb57ef,$(if $(filter $(ENV),staging),c45bad8691e1,$(if $(filter $(ENV),development),$(DEV_URL), "") ))
# Macro for production warning
PROD_WARNING = \
	if [ "$(ENV)" = "production" ]; then \
		echo "WARNING: You are about to execute this command on PRODUCTION!"; \
		read -p "Type 'PRODUCTION' to confirm: " CONFIRM && [ $$CONFIRM = "PRODUCTION" ] || (echo "Aborted!" && exit 1); \
	fi

set-env:
	@echo "Setting environment to $(env)"
	@echo "ENV=$(env)" > scripts/.env
	@echo ""
	@make status

status:
	@echo "Status:"
	@echo "Environment: $(ENV)"
	@echo "URL: $(URL)"
	@echo "APP: $(DOKPLOY_APP)"
	@echo "DB_PATH: $(DB_PATH)"
	@echo "DOCKER_CONTAINER: $(DOCKER_CONTAINER)"

download-db:
	@echo "Downloading with env: $(ENV)"
	scp root@142.93.224.213:$(DB_PATH) server_db/local_$(ENV).$$(date +%Y%m%d_%H%M).db

# Never tested, use with caution
upload-db:
	@$(PROD_WARNING)
	@echo "Uploading with env: $(ENV)"

	@echo "creating Backup"
	ssh root@142.93.224.213 "cp $(DB_PATH) $(DB_PATH).$$(date +%Y%m%d_%H%M%S).backup"

	scp server_db/${path} root@142.93.224.213:$(DB_PATH)

switch-db:
	@$(PROD_WARNING)
	ssh root@142.93.224.213 "mv $(DB_PATH) $(DB_PATH).$$(date +%Y%m%d_%H%M%S).unset.db && mv $(DB_PATH).$(BACKUP).backup $(DB_PATH)"

create-account:
	@$(PROD_WARNING)
	bash scripts/create_account.sh $(URL) $(NAME)

create-account-domain:
	@$(PROD_WARNING)
	bash scripts/create_domain.sh $(URL) $(ACCOUNT) $(DOMAIN)

delete-account-domain:
	@$(PROD_WARNING)
	bash scripts/delete_domain.sh $(URL) $(DOMAIN)

get-account:
	bash scripts/get_account.sh $(URL) $(ID)

get-user:
	bash scripts/get_user.sh $(URL) $(ID)

sqlc:
	cd db/sqlc && sqlc generate && cd ../../

migrate-goto:
	@$(PROD_WARNING)
	#echo "docker exec -it $(DOCKER_CONTAINER) migrate -path migrations -database sqlite3:///data/database.db?_foreign_keys=on goto $(number)"
	ssh -t root@142.93.224.213 "docker exec -it $(DOCKER_CONTAINER) migrate -path db/migrations -database sqlite3:///data/database.db?_foreign_keys=on goto $(number)"

migrate-down:
	@$(PROD_WARNING)
	#echo "docker exec -it $(DOCKER_CONTAINER) migrate -path migrations -database sqlite3:///data/database.db?_foreign_keys=on goto $(number)"
	ssh -t root@142.93.224.213 "docker exec -it $(DOCKER_CONTAINER) migrate -path db/migrations -database sqlite3:///data/database.db?_foreign_keys=on down $(number)"

migrate-up:
	@$(PROD_WARNING)
	#echo "docker exec -it $(DOCKER_CONTAINER) migrate -path migrations -database sqlite3:///data/database.db?_foreign_keys=on goto $(number)"
	ssh -t root@142.93.224.213 "docker exec -it $(DOCKER_CONTAINER) migrate -path db/migrations -database sqlite3:///data/database.db?_foreign_keys=on up $(number)"

migrate-version:
	@$(PROD_WARNING)
	#echo "docker exec -it $(DOCKER_CONTAINER) migrate -path migrations -database sqlite3:///data/database.db?_foreign_keys=on goto $(number)"
	ssh -t root@142.93.224.213 "docker exec -it $(DOCKER_CONTAINER) migrate -path db/migrations -database sqlite3:///data/database.db?_foreign_keys=on version"
migrate-force:
	@$(PROD_WARNING)
	#echo "docker exec -it $(DOCKER_CONTAINER) migrate -path migrations -database sqlite3:///data/database.db?_foreign_keys=on goto $(number)"
	ssh -t root@142.93.224.213 "docker exec -it $(DOCKER_CONTAINER) migrate -v -path db/migrations -database sqlite3:///data/database.db?_foreign_keys=on force $(number)"



#In case we messed something up:
#download-db-prod:
#	scp root@142.93.224.213:/etc/dokploy/compose/newproject-gochat-f51808/files/sqlite/database.db ./local_db.db

#backup-and-upload-staging:
#	ssh root@142.93.224.213 "cp /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db.$(date +%Y%m%d_%H%M%S).backup"
#	scp ./local_db.db root@142.93.224.213:/etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db
#
#backup-and-upload-db-production-torgon:
#	ssh root@142.93.224.213 "cp /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db.$(date +%Y%m%d_%H%M%S).backup"
#	scp ./local_db.db root@142.93.224.213:/etc/dokploy/compose/newproject-gochat-f51808/files/sqlite/database.db