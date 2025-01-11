windres -o .\Snorlax\Snorlax.syso .\Snorlax\Snorlax.rc
go generate .\Snorlax\
go build -ldflags "-H windowsgui" .\Snorlax\
go build -ldflags "-H windowsgui" .\update\