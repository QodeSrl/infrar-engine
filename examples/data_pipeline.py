"""
Data pipeline example using Infrar SDK
"""
from infrar.storage import upload, download, list_objects
import os

def process_data_files():
    """Process data files from cloud storage"""

    # Download input files
    input_files = list_objects(
        bucket='data-lake',
        prefix='raw/2024/'
    )

    for file_info in input_files:
        # Download file
        local_path = f'/tmp/{os.path.basename(file_info["key"])}'
        download(
            bucket='data-lake',
            source=file_info['key'],
            destination=local_path
        )

        # Process file (simplified)
        processed_path = process_file(local_path)

        # Upload processed file
        upload(
            bucket='data-lake',
            source=processed_path,
            destination=f'processed/2024/{os.path.basename(processed_path)}'
        )

def process_file(filepath):
    """Dummy processing function"""
    # In real scenario, this would transform the data
    processed_path = filepath.replace('.csv', '_processed.csv')
    # Simulate processing
    return processed_path

if __name__ == '__main__':
    process_data_files()
