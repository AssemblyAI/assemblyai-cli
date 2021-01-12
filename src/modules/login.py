from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from __future__ import unicode_literals
from modules.util import write_conf


def main(args):
    write_conf({'api_token': args.api_token})
