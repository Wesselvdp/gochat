hello:
	echo "Hello, world"

download-db-staging:
	scp root@142.93.224.213:/etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db ./local_db.db

download-db-prod:
	scp root@142.93.224.213:/etc/dokploy/compose/newproject-gochat-f51808/files/sqlite/database.db ./local_db.db

backup-and-upload-staging:
	ssh root@142.93.224.213 "cp /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db.$(date +%Y%m%d_%H%M%S).backup"
	scp ./local_db.db root@142.93.224.213:/etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db

backup-and-upload-db-production-torgon:
	ssh root@142.93.224.213 "cp /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db /etc/dokploy/compose/gochat-app-staging-bd0bf7/files/sqlite/database.db.$(date +%Y%m%d_%H%M%S).backup"
	scp ./local_db.db root@142.93.224.213:/etc/dokploy/compose/newproject-gochat-f51808/files/sqlite/database.db

