"""
Multiple storage operations using Infrar SDK
"""
from infrar.storage import upload, download, delete, list_objects

def manage_backups():
    """Manage backup files in cloud storage"""

    # Upload a file
    upload(
        bucket='my-backup-bucket',
        source='/tmp/report.pdf',
        destination='reports/monthly-report.pdf'
    )

    # List existing backups
    files = list_objects(
        bucket='my-backup-bucket',
        prefix='reports/'
    )
    print(f'Found {len(files)} files')

    # Download a backup
    download(
        bucket='my-backup-bucket',
        source='reports/monthly-report.pdf',
        destination='/tmp/downloaded-report.pdf'
    )

    # Delete old backup
    delete(
        bucket='my-backup-bucket',
        path='reports/old-report.pdf'
    )

if __name__ == '__main__':
    manage_backups()
