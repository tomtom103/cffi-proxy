import io

import httpx
import pandas as pd
from browserforge.headers import HeaderGenerator
import difflib
import pprint

def compare_dicts(d1, d2):
    return ('\n' + '\n'.join(difflib.ndiff(
                   pprint.pformat(d1).splitlines(),
                   pprint.pformat(d2).splitlines())))

def main() -> None:
    headers = HeaderGenerator()
    with httpx.Client(proxy="http://0.0.0.0:8000", verify=False) as client:
        # TODO: Figure out why randomizing headers changes the returned data type
        response = client.get("https://tls.browserleaks.com/tls", headers=headers.generate(http_version=1))
        dict1 = response.text
        print(dict1)

        response = client.get("https://tls.browserleaks.com/tls", headers=headers.generate(http_version=1))
        dict2 = response.json()

        compare_dicts(dict1, dict2)

if __name__ == "__main__":
    main()
