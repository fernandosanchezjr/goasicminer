import binascii
import unittest

from .reference_midstate import calculateMidstate


class TestMidstate(unittest.TestCase):

    def test_random_midstate(self):
        plain_text = 'b4076799ff727758f5b7ad0830d093671bffc34d5edbfa6afcc90820fd224111deb509351c2098173a0b99923e8bebf' \
                     '2b0e805de7cdc13b41ea33e6048abb056'
        mid_state = binascii.hexlify(calculateMidstate(binascii.unhexlify(plain_text)))
        assert mid_state == b'321885610e7c9a707dce85830158d5663f80a0fa8841afc73d469a80c4a70af4'
