#!/bin/bash
# ============================================================
# 自动备份数据库并压缩发送邮件
# 用法: ./backup-db.sh [收件邮箱]
# 定时任务: 0 3 * * * /path/to/backup-db.sh >> /var/log/db-backup.log 2>&1
# ============================================================

set -euo pipefail

# ---------- 配置 ----------
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
ENV_FILE="${SCRIPT_DIR}/backend-go/.env"
BACKUP_DIR="${SCRIPT_DIR}/backups"
KEEP_DAYS=7

RECIPIENT="${1:-leisureea@gmail.com}"
SENDER_NAME="知外助手数据库备份"

# ---------- 读取 .env ----------
if [ ! -f "$ENV_FILE" ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: .env not found at $ENV_FILE"
    exit 1
fi

get_env() { grep "^${1}=" "$ENV_FILE" | head -1 | cut -d'=' -f2-; }

DATABASE_URL=$(get_env "DATABASE_URL")

# 解析 DATABASE_URL: user:password@tcp(host:port)/dbname?params
DB_USER="${DATABASE_URL%%:*}"
DB_REST="${DATABASE_URL#*:}"          # password@tcp(host:port)/dbname?params
DB_PASS="${DB_REST%%@*}"
DB_TCP="${DB_REST#*@tcp(}"           # host:port)/dbname?params
DB_ADDR="${DB_TCP%%)*}"              # host:port
DB_HOST="${DB_ADDR%%:*}"
DB_PORT="${DB_ADDR#*:}"
DB_PORT="${DB_PORT:-3306}"
DB_PATH="${DB_TCP#*/}"               # dbname?params
DB_NAME="${DB_PATH%%\?*}"

MAIL_HOST=$(get_env "MAIL_HOST")
MAIL_PORT=$(get_env "MAIL_PORT")
MAIL_USER=$(get_env "MAIL_USERNAME")
MAIL_PASS=$(get_env "MAIL_PASSWORD")
MAIL_FROM=$(get_env "MAIL_FROM")

for var in DB_USER DB_PASS DB_HOST DB_NAME MAIL_HOST MAIL_PORT MAIL_USER MAIL_PASS MAIL_FROM; do
    if [ -z "${!var:-}" ]; then
        echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: $var is empty in .env"
        exit 1
    fi
done

# ---------- 准备备份目录 ----------
mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date '+%Y%m%d_%H%M%S')
BACKUP_FILE="${BACKUP_DIR}/${DB_NAME}_${TIMESTAMP}.sql"
COMPRESSED_FILE="${BACKUP_FILE}.gz"

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Starting database backup: $DB_NAME@$DB_HOST"

# ---------- 导出数据库 ----------
if ! mysqldump -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASS" \
    --single-transaction --routines --triggers --events \
    --default-character-set=utf8mb4 \
    "$DB_NAME" > "$BACKUP_FILE" 2>/dev/null; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] ERROR: mysqldump failed"
    rm -f "$BACKUP_FILE"
    exit 1
fi

# ---------- 压缩 ----------
gzip -f "$BACKUP_FILE"
FILE_SIZE=$(du -h "$COMPRESSED_FILE" | cut -f1)
echo "[$(date '+%Y-%m-%d %H:%M:%S')] Backup completed: $COMPRESSED_FILE ($FILE_SIZE)"

# ---------- 通过 Python 发送邮件 ----------
export DB_NAME DB_HOST COMPRESSED_FILE FILE_SIZE \
       MAIL_HOST MAIL_PORT MAIL_USER MAIL_PASS MAIL_FROM RECIPIENT SENDER_NAME

python3 - <<PYEOF
import smtplib, ssl, os
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from email.mime.application import MIMEApplication
from datetime import datetime

subject = f"【{os.environ['SENDER_NAME']}】{os.environ['DB_NAME']} {datetime.now().strftime('%Y-%m-%d %H:%M')}"
attachment = os.environ['COMPRESSED_FILE']
file_size = os.environ['FILE_SIZE']

msg = MIMEMultipart()
msg['From'] = f"{os.environ['SENDER_NAME']} <{os.environ['MAIL_FROM']}>"
msg['To'] = os.environ['RECIPIENT']
msg['Subject'] = subject

body = f"""数据库备份已完成。

  数据库: {os.environ['DB_NAME']}
  主机: {os.environ['DB_HOST']}
  时间: {datetime.now().strftime('%Y-%m-%d %H:%M:%S')}
  大小: {file_size}

此邮件由自动备份脚本发送。
"""
msg.attach(MIMEText(body, 'plain', 'utf-8'))

with open(attachment, 'rb') as f:
    part = MIMEApplication(f.read(), Name=os.path.basename(attachment))
    part['Content-Disposition'] = f'attachment; filename="{os.path.basename(attachment)}"'
    msg.attach(part)

context = ssl.create_default_context()
try:
    with smtplib.SMTP_SSL(os.environ['MAIL_HOST'], int(os.environ['MAIL_PORT']), context=context, timeout=60) as server:
        server.login(os.environ['MAIL_USER'], os.environ['MAIL_PASS'])
        server.send_message(msg)
    print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] Email sent to {os.environ['RECIPIENT']} ({file_size})")
except Exception as e:
    print(f"[{datetime.now().strftime('%Y-%m-%d %H:%M:%S')}] ERROR: Failed to send email: {e}")
    exit(1)
PYEOF

# ---------- 清理旧备份 ----------
DELETED=$(find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime +$KEEP_DAYS -delete -print 2>/dev/null | wc -l)
if [ "$DELETED" -gt 0 ]; then
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] Cleaned up $DELETED backup(s) older than $KEEP_DAYS days"
fi

echo "[$(date '+%Y-%m-%d %H:%M:%S')] Done."
