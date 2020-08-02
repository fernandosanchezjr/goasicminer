import binascii
import hashlib
import json
import typing as t
import unittest

# Inspired by details from https://slushpool.com/help/stratum-protocol/

PADDING = '000000800000000000000000000000000000000000000000000000000000000000000000000000000000000080020000'


def double_sha(value: t.Union[str, bytes]) -> bytes:
    return binascii.hexlify(hashlib.sha256(hashlib.sha256(binascii.unhexlify(value)).digest()).digest())


class WorkToHeaderTest(unittest.TestCase):

    def setUp(self):
        with open('subscribe_test.json') as f:
            self.subscribe_data = json.load(f)
        with open('set_difficulty_test.json') as f:
            self.set_difficulty = json.load(f)
        with open('notify_test.json') as f:
            self.notify = json.load(f)

    def generate_coinbase(self) -> bytes:
        # dumb way of zero padding
        extraNonce1 = ('0' * (16 - len(self.subscribe_data['result'][1]))) + self.subscribe_data['result'][1]
        extraNonce2 = '0000000000000000'
        coinb1 = self.notify['params'][2]
        coinb2 = self.notify['params'][3]
        plain_coinbase = coinb1 + extraNonce1 + extraNonce2 + coinb2
        assert plain_coinbase == '01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4' \
                                 'a031ccb09fabe6d6df183ff6cbf2a1e8198b6679b7cef3e1cce0431da353154caa55e04ca3f66b3a601' \
                                 '00000000000000002a6502002aa65f0000000000000000939d289b2f736c7573682f000000000443ca3' \
                                 'c26000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac00000000000000002c6a' \
                                 '4c2952534b424c4f434b3a0ec82b00b353ab052014b472cb3ee39bb32431be99b7db757171f71600275' \
                                 '0190000000000000000296a4c266a24b9e11b6d8f8cc50f47dc5e8537a9e300984ee50eefd8eb7917b4' \
                                 'b83a28287fe15e80c9820000000000000000266a24aa21a9ed7aee68d448839eba918f66147bf31b096' \
                                 'fe443c60175a53878c8e052cdd799f700000000'
        return double_sha(plain_coinbase)

    def test_coinbase_hash(self):
        coinbase_hash = self.generate_coinbase()
        assert coinbase_hash == b'2d749fc9eeea345bd91241187f92318442f48fca3c2537c242d2c6c917d7dca6'

    def generate_merkle_root(self) -> bytes:
        coinbase_hash = self.generate_coinbase()
        merkle_branches = self.notify['params'][4]
        merkle_root = coinbase_hash
        for branch in merkle_branches:
            merkle_root = double_sha(merkle_root + branch.encode())
        return merkle_root

    def test_merkle_root(self):
        merkle_root = self.generate_merkle_root()
        assert merkle_root == b'8de8f457cffef502d75ada232b2e68be61724c35f48432c7d0cac77d7b1dde50'

    def test_block_header(self):
        version = self.notify['params'][5]
        prev_hash = self.notify['params'][1]
        merkle_root = self.generate_merkle_root()
        nbits = self.notify['params'][6]
        ntime = self.notify['params'][7]
        nonce = '00000000'

        plain_header = version.encode() + prev_hash.encode() + merkle_root + ntime.encode() + nbits.encode() + \
                       nonce.encode() + PADDING.encode()
        header_hash = double_sha(plain_header)

        assert plain_header == b'20000000bd3e4f2c6d8b14c9d677cb428a124dcae58c5530000f791f00000000000000008de8f457cffe' \
                               b'f502d75ada232b2e68be61724c35f48432c7d0cac77d7b1dde505f2606591710b4f80000000000000080' \
                               b'000000000000000000000000000000000000000000000000000000000000000000000000000000008002' \
                               b'0000'
        assert header_hash == b'c0554499cb6404341b6cec408e778ca5fc96c8c3612d5005edf699f4179d8392'


if __name__ == '__main__':
    unittest.main(verbosity=2)
