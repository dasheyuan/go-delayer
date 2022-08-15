@echo off

SET ServiceName=%2
SET Version=%3"
SET HarborUrl=%1

echo ----- Step 1: build -----
go env -w CGO_ENABLED=0
go env -w GOOS=linux
kratos build
go env -w CGO_ENABLED=1
go env -w GOOS=windows
echo.

echo ----- Step 2: build docker images -----
docker build -t %ServiceName% .
echo.

echo ----- Step 3: tag images -----
docker tag %ServiceName% %HarborUrl%/%ServiceName%:%Version%
docker tag %ServiceName% %HarborUrl%/%ServiceName%:latest
echo.

echo ----- Step 4: push images -----
docker push %HarborUrl%/%ServiceName%:%Version%
docker push %HarborUrl%/%ServiceName%:latest
echo.

echo ----- Step 5: remove images -----
docker rmi %HarborUrl%/%ServiceName%:%Version%
docker rmi %HarborUrl%/%ServiceName%:latest
docker rmi %ServiceName%
echo.

echo ----- Step 6: finish -----
pause