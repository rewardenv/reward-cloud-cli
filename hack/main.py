import sys
import argparse
from pathlib import Path
import ruamel.yaml

parser = argparse.ArgumentParser()
parser.add_argument("-i", "--input", help="input file in yaml format", required=True, type=argparse.FileType('r'))
parser.add_argument("-o", "--output", help="output file in yaml format, defaults to output.yaml", default="reward-api.yaml", type=argparse.FileType('w'))

args = parser.parse_args()

in_file = args.input
out_file = args.output

yaml = ruamel.yaml.YAML()
yaml.indent(mapping=2, sequence=4, offset=2)
yaml.preserve_quotes = True
data = yaml.load(in_file)

delete = []


def rm_merge(d, s):
    if isinstance(d, dict):
        if s in d:
            del d[s]
        for k in list(d.keys()):
            if str(k).find(s) != -1:
                del d[k]
            else:
                rm_merge(d[k], s)
    elif isinstance(d, list):
        for item in d:
            rm_merge(item, s)


rm_merge(data, 'application/ld+json')
rm_merge(data, '.jsonld')
rm_merge(data, 'application/jsonld')

yaml.dump(data, out_file)
