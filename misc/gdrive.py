from __future__ import print_function
import sys
import io
import pip
import httplib2
import os
from mimetypes import MimeTypes

import simplejson
from apiclient import errors


try:
    from googleapiclient.errors import HttpError
    from apiclient import discovery
    import oauth2client
    from oauth2client import client
    from oauth2client import tools
except ImportError:
    print('goole-api-python-client is not installed. Try:')
    print('sudo pip install --upgrade google-api-python-client')
    sys.exit(1)


class Flag:
    auth_host_name = 'localhost'
    noauth_local_webserver = False
    auth_host_port = [8080, 8090]
    logging_level = 'ERROR'


try:
    import argparse

    flags = Flag()
except ImportError:
    flags = None

# If modifying these scopes, delete your previously saved credentials
# at ~/.credentials/drive-python-quickstart.json
SCOPES = 'https://www.googleapis.com/auth/drive'
CLIENT_SECRET_FILE = 'config.json'
APPLICATION_NAME = 'GDrive'


def get_credentials():

    home_dir = os.path.expanduser('~')
    credential_dir = os.path.join(home_dir, '.credentials')
    if not os.path.exists(credential_dir):
        os.makedirs(credential_dir)
    credential_path = os.path.join(credential_dir,
                                   'drive-python-quickstart.json')

    store = oauth2client.file.Storage(credential_path)
    credentials = store.get()
    if not credentials or credentials.invalid:
        flow = client.flow_from_clientsecrets(CLIENT_SECRET_FILE, SCOPES)
        flow.user_agent = APPLICATION_NAME
        # if flags:
        credentials = tools.run_flow(flow, store, flags)
        # else:  # Needed only for compatibility with Python 2.6
        #     credentials = tools.run(flow, store)
        print('Storing credentials to ' + credential_path)
    return credentials


def list_files(parent):
    page_token = None
    while True:
        try:
            results = service.files().list(
                q="'%s' in parents and trashed=false" %
                parent,
                fields='nextPageToken, files(mimeType, id, md5Checksum)',
                pageToken=page_token).execute()
            items = results.get('files', [])

            for i in items:
                if i['mimeType'] == 'application/vnd.google-apps.folder':
                    list_files(i['id'])
                else:
                    print('{0} {1}'.format(i['id'], i['md5Checksum']))
            page_token = results.get('nextPageToken', None)
            if page_token is None:
                break

        except(errors.HttpError, e):
            try:
                # Load Json body.
                error = simplejson.loads(e.content)
                print('Error code: %d' % error.get('code'))
                print('Error message: %s' % error.get('message'))
                # More error information can be retrieved with
                # error.get('errors').
            except ValueError:
                # Could not load Json body.
                print('HTTP Status code: %d' % e.resp.status)
                print('HTTP Reason: %s' % e.resp.reason)


if __name__ == '__main__':
    credentials = get_credentials()
    http = credentials.authorize(httplib2.Http())
    service = discovery.build('drive', 'v3', http=http)

    method = sys.argv[1]
    if method == 'list':
        list_files(sys.argv[2])
