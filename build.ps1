windres -o Snorlax.syso Snorlax.rc
go generate
go build -ldflags "-H windowsgui"