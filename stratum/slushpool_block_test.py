import binascii
import hashlib
import json
import typing as t
import unittest

# Inspired by details from https://slushpool.com/help/stratum-protocol/

PADDING = '000000800000000000000000000000000000000000000000000000000000000000000000000000000000000080020000'
SUBSCRIBE_JSON = '{"id":1,"method":"","params":null,"result":[[["mining.set_difficulty","1"],["mining.notify","1"]],' \
                 '"2a6502002aa65f",8],"error":null}'
SET_DIFFICULTY_JSON = '{"id":0,"method":"mining.set_difficulty","params":[8192],"result":null,"error":null}'
NOTIFY_JSON = '{"id":0,"method":"mining.notify","params":["9b289d93","bd3e4f2c6d8b14c9d677cb428a124dcae58c5530000f791' \
              'f0000000000000000","01000000010000000000000000000000000000000000000000000000000000000000000000ffffffff' \
              '4a031ccb09fabe6d6df183ff6cbf2a1e8198b6679b7cef3e1cce0431da353154caa55e04ca3f66b3a60100000000000000","9' \
              '39d289b2f736c7573682f000000000443ca3c26000000001976a9147c154ed1dc59609e3d26abb2df2ea3d587cd8c4188ac000' \
              '00000000000002c6a4c2952534b424c4f434b3a0ec82b00b353ab052014b472cb3ee39bb32431be99b7db757171f7160027501' \
              '90000000000000000296a4c266a24b9e11b6d8f8cc50f47dc5e8537a9e300984ee50eefd8eb7917b4b83a28287fe15e80c9820' \
              '000000000000000266a24aa21a9ed7aee68d448839eba918f66147bf31b096fe443c60175a53878c8e052cdd799f700000000"' \
              ',["f3dbd0071549db620a9e0969e54d9bec3d22093817b48e7d1cb0e02edbf698d4","c6e9ebbf95ac8ec33d0349f6a05b71e4' \
              '28a4bf3ce74e1e2b1774504bc1f68f39","be214bbc6c1b82ddc69c440bdfaaa63e5be2760cc9327868f0336050f77b0928","' \
              '08d41fb5297c248b58682ea7e967ddbd712bd3985113a2c556a57ef464b604fc","c95ef4a3995bfe6728e7214a720851d4d14' \
              'ebee3518d2d4aa6b026f44e16413d","1a589249b64cae830be976d537ee5d7cf9776cc8589d5c83d0dbad909f0fa002","390' \
              '3aeea64743ca14bb2fae0f82d61b116f6b048d063e978102e2a1c015b6501","9581e992663090ced5c5cbbef8cd95028f804c' \
              '45a6361ecff9ec699b31922573","7de724b8c18fa8030979d89aa3adc5dc08f59132005b1b39173e05bf88b4fe3e","b309c1' \
              '57d95cf434627e7f0ea5d87d4cbf287cb4a06fbc55609db4e3b7bbe3d0","77f11484dbd99409257933d3da53b9b7422232393' \
              '78f290bf051a2ee4521fde0"],"20000000","1710b4f8","5f260659",true],"result":null,"error":null}'


def double_sha(value: t.Union[str, bytes]) -> bytes:
    return binascii.hexlify(hashlib.sha256(hashlib.sha256(binascii.unhexlify(value)).digest()).digest())


class WorkToHeaderTest(unittest.TestCase):

    def setUp(self):
        self.subscribe_data = json.loads(SUBSCRIBE_JSON)
        self.set_difficulty = json.loads(SET_DIFFICULTY_JSON)
        self.notify = json.loads(NOTIFY_JSON)

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
