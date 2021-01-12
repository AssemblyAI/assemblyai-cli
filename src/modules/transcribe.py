from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from __future__ import unicode_literals
from modules.util import read_conf
import requests
import json
import time
import sys


POLLING_INTERVAL_SECONDS = 1


def read_file(filename, chunk_size=5242880):
    with open(filename, 'rb') as _file:
        while True:
            data = _file.read(chunk_size)
            if not data:
                break
            yield data


def upload_file(filename, api_token):
    headers = {'authorization': api_token}
    response = requests.post('https://api.assemblyai.com/v2/upload',
                             headers=headers,
                             data=read_file(filename))
    response.raise_for_status()
    return response.json()['upload_url']


def poll_status(transcript_id, api_token):
    endpoint = "https://api.assemblyai.com/v2/transcript/{transcript_id}".format(
        transcript_id=transcript_id
    )

    headers = {
        "authorization": api_token,
    }

    return requests.get(endpoint, headers=headers).json()


def main(args):
    conf = read_conf()
    api_token = conf['api_token']
    endpoint = "https://api.assemblyai.com/v2/transcript"
    error = None
    audio_url = None
    response = {}

    if args.audio_url is None:
        # Use upload API
        try:
            audio_url = upload_file(args.audio_file, api_token)
        except Exception as e:
            error = 'Error uploading file: {}'.format(e)
    else:
        # Use audio url directly
        audio_url = args.audio_url

    if error is None:
        data = {
            "audio_url": audio_url
        }

        headers = {
            "authorization": api_token,
            "content-type": "application/json"
        }

        response = requests.post(endpoint, json=data, headers=headers)\

        try:
            response.raise_for_status()
        except Exception as e:
            error = 'Transcribe API returned a non 200 status code: {}'.format(e)

        if not error:
            response = response.json()
            transcript_id = response['id']
            status = response['status']
            while status != 'completed':
                if not args.quiet:
                    print('Status: {}'.format(status))
                time.sleep(POLLING_INTERVAL_SECONDS)
                response = poll_status(transcript_id, api_token)
                status = response['status']
                if status == 'error':
                    error = response['error']
                    break

            if args.text_only:
                response = response['text']

    if error:
        response = {
            'error': error
        }

    if args.output_to:
        with open(args.output_to, 'w') as f:
            f.write(json.dumps(response, indent=4))
        print('Completed. Results written to {}.'.format(args.output_to))
    else:
        print(json.dumps(response, indent=4))

    if error:
        sys.exit(1)
