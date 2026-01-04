#!/usr/bin/env python3
"""
Jira Attachment Downloader

This script downloads attachments from Jira tickets for both Malaysia (mybuddy)
and Singapore (sgbuddy) environments.

Usage:
    python download_jira_attachment.py TICKET-ID [options]

Examples:
    # Download all attachments from a ticket (auto-detect region)
    python download_jira_attachment.py TS-1234

    # Download specific attachment by name
    python download_jira_attachment.py TS-1234 --filename transactions.csv

    # Download from Singapore region
    python download_jira_attachment.py TSE-567 --region sg

    # List attachments without downloading
    python download_jira_attachment.py TS-1234 --list-only

    # Download to specific directory
    python download_jira_attachment.py TS-1234 --output ./downloads
"""

import argparse
import os
import sys
from pathlib import Path
from typing import List, Dict, Optional
import requests
from urllib.parse import urljoin


class JiraAttachmentDownloader:
    """Downloads attachments from Jira tickets"""

    # Jira domains
    DOMAINS = {
        'my': 'https://gxbank.atlassian.net',
        'sg': 'https://gxsbank.atlassian.net'
    }

    def __init__(self, region: str = 'my'):
        """
        Initialize the downloader

        Args:
            region: 'my' for Malaysia or 'sg' for Singapore
        """
        if region not in self.DOMAINS:
            raise ValueError(f"Invalid region: {region}. Must be 'my' or 'sg'")

        self.region = region
        self.domain = self.DOMAINS[region]

        # Get credentials from environment
        self.username = os.environ.get('JIRA_USERNAME')
        self.api_key = os.environ.get('JIRA_API_KEY')

        if not self.username or not self.api_key:
            raise ValueError(
                "JIRA_USERNAME and JIRA_API_KEY environment variables must be set"
            )

        self.auth = (self.username, self.api_key)
        self.api_base = f"{self.domain}/rest/api/3"

    def get_ticket_attachments(self, ticket_key: str) -> List[Dict]:
        """
        Get all attachments for a ticket

        Args:
            ticket_key: Jira ticket key (e.g., TS-1234)

        Returns:
            List of attachment dictionaries
        """
        url = f"{self.api_base}/issue/{ticket_key}?fields=attachment"

        response = requests.get(url, auth=self.auth)
        response.raise_for_status()

        data = response.json()
        attachments = []

        for att in data.get('fields', {}).get('attachment', []):
            attachments.append({
                'id': att['id'],
                'filename': att['filename'],
                'mimeType': att['mimeType'],
                'url': att['content'],
                'size': att.get('size', 0)
            })

        return attachments

    def download_attachment(self, attachment_url: str, output_path: str) -> bool:
        """
        Download a single attachment

        Args:
            attachment_url: URL to download from
            output_path: Where to save the file

        Returns:
            True if successful, False otherwise
        """
        try:
            response = requests.get(attachment_url, auth=self.auth, stream=True)
            response.raise_for_status()

            # Handle duplicate filenames
            final_path = self._resolve_path(output_path)

            with open(final_path, 'wb') as f:
                for chunk in response.iter_content(chunk_size=8192):
                    if chunk:
                        f.write(chunk)

            print(f"✓ Downloaded: {final_path}")
            return True

        except Exception as e:
            print(f"✗ Failed to download {output_path}: {e}", file=sys.stderr)
            return False

    def _resolve_path(self, path: str) -> str:
        """
        Resolve final save path, handling duplicates

        Args:
            path: Desired save path

        Returns:
            Final path (with suffix if file exists)
        """
        if not os.path.exists(path):
            return path

        # File exists - add suffix
        base = Path(path)
        counter = 1

        while True:
            new_path = base.parent / f"{base.stem}_{counter}{base.suffix}"
            if not new_path.exists():
                return str(new_path)
            counter += 1

    def download_ticket_attachments(
        self,
        ticket_key: str,
        output_dir: str = ".",
        filename_filter: Optional[str] = None
    ) -> int:
        """
        Download all (or filtered) attachments from a ticket

        Args:
            ticket_key: Jira ticket key (e.g., TS-1234)
            output_dir: Directory to save files to
            filename_filter: Optional filename filter (e.g., "transactions.csv")

        Returns:
            Number of successfully downloaded files
        """
        print(f"Fetching attachments for {ticket_key} from {self.domain}...")
        attachments = self.get_ticket_attachments(ticket_key)

        if not attachments:
            print("No attachments found.")
            return 0

        print(f"\nFound {len(attachments)} attachment(s):")
        for att in attachments:
            size_mb = att['size'] / (1024 * 1024) if att['size'] > 0 else 0
            print(f"  - {att['filename']} ({att['mimeType']}, {size_mb:.2f} MB)")

        # Filter by filename if specified
        if filename_filter:
            attachments = [a for a in attachments if a['filename'] == filename_filter]
            print(f"\nFiltered to {len(attachments)} file(s) matching '{filename_filter}'")

        # Create output directory if needed
        os.makedirs(output_dir, exist_ok=True)

        # Download attachments
        print(f"\nDownloading to: {output_dir}/")
        success_count = 0

        for att in attachments:
            output_path = os.path.join(output_dir, att['filename'])
            if self.download_attachment(att['url'], output_path):
                success_count += 1

        print(f"\nSuccessfully downloaded {success_count}/{len(attachments)} file(s)")
        return success_count

    def list_attachments(self, ticket_key: str) -> None:
        """
        List attachments without downloading

        Args:
            ticket_key: Jira ticket key
        """
        attachments = self.get_ticket_attachments(ticket_key)

        if not attachments:
            print("No attachments found.")
            return

        print(f"\nAttachments for {ticket_key}:")
        print("-" * 80)

        for att in attachments:
            size_mb = att['size'] / (1024 * 1024) if att['size'] > 0 else 0
            print(f"Filename: {att['filename']}")
            print(f"  Type: {att['mimeType']}")
            print(f"  Size: {size_mb:.2f} MB ({att['size']} bytes)")
            print(f"  URL: {att['url']}")
            print()


def detect_region_from_ticket(ticket_key: str) -> str:
    """
    Auto-detect region from ticket key prefix

    Args:
        ticket_key: Jira ticket key

    Returns:
        'my' or 'sg'
    """
    if ticket_key.upper().startswith('TSE'):
        return 'sg'
    elif ticket_key.upper().startswith('TS'):
        return 'my'
    else:
        # Default to mybuddy for unknown formats
        print(f"Warning: Unknown ticket format '{ticket_key}', defaulting to Malaysia (my)")
        return 'my'


def main():
    parser = argparse.ArgumentParser(
        description='Download attachments from Jira tickets',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog=__doc__
    )

    parser.add_argument(
        'ticket',
        help='Jira ticket key (e.g., TS-1234, TSE-567)'
    )

    parser.add_argument(
        '-r', '--region',
        choices=['my', 'sg'],
        help='Region (my=Malaysia, sg=Singapore). Auto-detected from ticket key if not specified.'
    )

    parser.add_argument(
        '-f', '--filename',
        help='Download only attachment matching this filename (e.g., transactions.csv)'
    )

    parser.add_argument(
        '-o', '--output',
        default='.',
        help='Output directory (default: current directory)'
    )

    parser.add_argument(
        '-l', '--list-only',
        action='store_true',
        help='List attachments without downloading'
    )

    args = parser.parse_args()

    # Detect region if not specified
    region = args.region or detect_region_from_ticket(args.ticket)

    try:
        downloader = JiraAttachmentDownloader(region=region)

        if args.list_only:
            downloader.list_attachments(args.ticket)
        else:
            downloader.download_ticket_attachments(
                ticket_key=args.ticket,
                output_dir=args.output,
                filename_filter=args.filename
            )

    except ValueError as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)
    except requests.exceptions.RequestException as e:
        print(f"Network error: {e}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nDownload cancelled.")
        sys.exit(1)


if __name__ == '__main__':
    main()
