"""
Simple file upload example using Infrar SDK
"""
from infrar.storage import upload

def backup_file():
    """Backup a single file to cloud storage"""
    upload(
        bucket='my-backup-bucket',
        source='/tmp/data.csv',
        destination='backups/2024/data.csv'
    )
    print('File uploaded successfully!')

if __name__ == '__main__':
    backup_file()
