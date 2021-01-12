from __future__ import absolute_import
from __future__ import division
from __future__ import print_function
from __future__ import unicode_literals
# END PYTHON2 COMPATIBILITY
import argparse
import sys
from modules.feedback import main as feedback
from modules.login import main as login
from modules.transcribe import main as transcribe


def validate_arguments(args):
    if args.method == 'login':
        if args.api_token is None:
            raise RuntimeError("Login method requires API Token")
    elif args.method == 'feedback':
        if args.message is None:
            raise RuntimeError("Feedback method requires Message (--message) parameter")
    elif args.method == 'transcribe':
        if args.audio_file is None and args.audio_url is None:
            raise RuntimeError("Transcribe method requires one of --audio_file or --audio_url parameters")
        elif args.audio_file is not None and args.audio_url is not None:
            raise RuntimeError("Transcribe method only supports one of --audio_file or --audio_url parameters (not both)")


def main():

    # Parse command line arguments
    parser = argparse.ArgumentParser(description="AssemblyAI CLI")
    parser.add_argument('method', help='CLI Command to perform (login, help, feedback, transcribe).',
                        choices=['login', 'help', 'feedback', 'transcribe'])

    # Arguments for "login"
    parser.add_argument('--api_token', help='API Token to authenticate with AssemblyAI.',
                        required=False)

    # Arguments for "feedback"
    parser.add_argument('--message', help='Feedback to send', required=False)

    # Arguments for "transcribe"
    parser.add_argument('--audio_file',
                        help='Path to local audio file (https://docs.assemblyai.com/guides/uploading-audio-files-for-transcription)',
                        required=False)
    parser.add_argument('--audio_url',
                        help='URL of remote audio file (https://docs.assemblyai.com/guides/transcribing-an-audio-file-recording)',
                        required=False)
    parser.add_argument('--output_to',
                        help='File to output the transcription results (prints to stdout if blank)',
                        required=False)
    parser.add_argument('--quiet',
                        help='If present, CLI will suppress status messages',
                        action='store_true')
    parser.add_argument('--text_only',
                        help='If present, only the text of the transcription will be returned',
                        action='store_true')
    parser.add_argument('--word_boost',
                        help='List of words to boost accuracy for; eg: "foo,bar,foo bar" *optional* (https://docs.assemblyai.com/guides/boosting-accuracy-for-keywords-or-phrases)',
                        required=False)
    parser.add_argument('--boost_weight',
                        choices=["low","default","high"],
                        help='Boost weight to use *optional* (https://docs.assemblyai.com/guides/boosting-accuracy-for-keywords-or-phrases)',
                        required=False)
    parser.add_argument('--model',
                        choices=["assemblyai_default", "assemblyai_en_au", "assemblyai_en_uk"],
                        help='Model to use *optional* (https://docs.assemblyai.com/guides/transcribing-with-a-different-acoustic-or-custom-language-model)',
                        required=False)

    args = parser.parse_args()

    try:
        validate_arguments(args)
    except Exception as e:
        print('Error: {}'.format(e))
        parser.print_usage()
        sys.exit(1)

    method = args.method

    if method == 'help':
        parser.print_help()
    elif method == 'login':
        login(args)
    elif method == 'feedback':
        feedback(args)
    elif method == 'transcribe':
        transcribe(args)


if __name__ == '__main__':
    main()