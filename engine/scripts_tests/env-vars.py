#!/usr/bin/env python3

import json
import os

content = {
    k: v for k, v in os.environ.items()
}

print(json.dumps(content))
