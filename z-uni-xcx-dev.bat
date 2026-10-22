@echo off
chcp 65001 >nul

:: 切换到当前批处理文件所在的目录
cd uni

:: 使用 call 强制批处理等待命令执行，防止直接闪退
call pnpm run dev:mp-weixin


pause