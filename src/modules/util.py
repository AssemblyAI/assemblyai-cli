from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from __future__ import unicode_literals
from os.path import expanduser
import os
import json

home_dir = expanduser("~")
assemblyai_home_dir = os.path.join(home_dir, '.assemblyai')


def write_conf(conf):
    if not os.path.exists(assemblyai_home_dir):
        os.makedirs(assemblyai_home_dir)
    assemblyai_conf_file = os.path.join(assemblyai_home_dir, 'conf')
    with open(assemblyai_conf_file, 'w') as f:
        f.write(json.dumps(conf, indent=4))


def read_conf():
    assemblyai_conf_file = os.path.join(assemblyai_home_dir, 'conf')
    if not os.path.isfile(assemblyai_conf_file):
        raise RuntimeError("Unable to find AssemblyAI config. Please login using the CLI.")
    with open(assemblyai_conf_file, 'r') as f:
        return json.loads(f.read().strip())
