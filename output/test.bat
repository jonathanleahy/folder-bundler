@REM del folder-bundler.exe
@REM go build ..
@REM
@REM rmdir inputtest /s /q
@REM xcopy source inputtest /E /I
@REM del inputtest_collated_part1.md
@REM folder-bundler.exe collect inputtest
@REM rmdir inputtest /s /q
@REM folder-bundler.exe reconstruct inputtest_collated_part1.md

del folder-bundler.exe
go build ..

rmdir temp /s /q
@REM xcopy source inputtest /E /I
@REM del inputtest_collated_part1.md
@REM folder-bundler.exe collect inputtest
@REM rmdir inputtest /s /q
folder-bundler.exe reconstruct temp.md