module CodeDocumentsArchiver

go 1.21.4

replace github.com/motaz/codeutils => /home/motaz/CodeProjects/Utils/codeutils

require (
	github.com/go-sql-driver/mysql v1.6.0
	//	github.com/motaz/codeutils v1.0.25
	github.com/motaz/redisaccess v1.0.3

)

require github.com/motaz/codeutils v0.0.0-00010101000000-000000000000

require github.com/go-redis/redis/v7 v7.4.1 // indirect
