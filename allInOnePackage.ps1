
Remove-Item -Path ./release -Force -Recurse
New-Item -Path ./release -ItemType Directory
New-Item -Path ./release/web -ItemType Directory
New-Item -Path ./release/conf -ItemType Directory

Copy-Item -Path ./conf/app.conf -Destination ./release/conf


# 复制配置文件
Copy-Item -Path ./conf/app.conf -Destination ./release/conf

# 修改配置文件，指明前端编译产物路径
$confPath = "./release/conf/app.conf"
$tempFile = "$env:TEMP\temp_config.conf"

# 逐行处理并修改特定键
Get-Content $confPath | ForEach-Object {
    if ($_ -match '^\s*staticBaseUrl\s*=') {
        'staticBaseUrl = "./web/build"'
    } else {
        $_
    }
} | Set-Content $tempFile -Encoding UTF8

# 替换原文件
Move-Item $tempFile $confPath -Force


# 编译前端
cd web
npm run build
cd ../
# 将前端产物复制到release/web
Copy-Item -Path ./web/build -Destination ./release/web -Recurse -Force

# 编译后端
$env:CGO_ENABLED = "0"
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags="-w -s" -o ./release/casdoor-Linux

# $env:CGO_ENABLED = "0"
# $env:GOOS = "windows"
# $env:GOARCH = "amd64"
# go build -ldflags="-w -s" -o ./release/casdoor-WIN.exe


cd release

tar -czvf "casdoor.tar.gz" .


