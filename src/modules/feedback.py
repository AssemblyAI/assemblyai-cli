from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from __future__ import unicode_literals
import requests
from modules.util import read_conf


def main(args):
    endpoint = "https://api.assemblyai.com/internal/cli/feedback"
    conf = read_conf()
    api_token = conf['api_token']

    message = args.message

    headers = {
        "authorization": api_token
    }

    requests.post(endpoint, headers=headers, json={'message': message})

    if not args.quiet:
        print('Feedback sent.')