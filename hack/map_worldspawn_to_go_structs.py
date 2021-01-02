#!/usr/bin/env python3
"""
converts first entity in map file to out quakemap.Brush declaration in golang.
cheesed by using python's own tokenizer
"""

import token
import tokenize
import argparse
import sys
from pprint import pprint


class IgnoreCommentsTokenizer:
    """
    Ignores single-line comments
    """
    def __init__(self, freadline):
        self.tokenizer = tokenize.generate_tokens(freadline)
        self.cur_token_type = None
        self.cur_token_string = None
        self.next_token_ignore_comments()

    def next_token(self):
        self.cur_token_type, self.cur_token_string, _, _, _ = \
                self.tokenizer.__next__()

    def next_token_ignore_comments(self):
        self.next_token()
        while self.cur_token_type == token.OP and self.cur_token_string == '//':
            while self.cur_token_type not in (token.NEWLINE, token.NL):
                self.next_token()
            self.next_token()

    def next(self):
        cur_token_type = self.cur_token_type
        cur_token_string = self.cur_token_string
        if self.cur_token_type != token.ENDMARKER:
            self.next_token_ignore_comments()
        return cur_token_type, cur_token_string

    def peek(self):
        return self.cur_token_type, self.cur_token_string


class ParseError(Exception):
    pass


class Plane:
    def __init__(self):
        self.x1 = 0.0
        self.y1 = 0.0
        self.z1 = 0.0
        self.x2 = 0.0
        self.y2 = 0.0
        self.z2 = 0.0
        self.x3 = 0.0
        self.y3 = 0.0
        self.z3 = 0.0
        self.texture = None
        self.xoffset = self.yoffset = 0.0
        self.rotation = 0.0
        self.xscale = self.yscale = 0.0


class Brush:
    def __init__(self):
        self.planes = []


class Entity:
    def __init__(self):
        self.keys = {}
        self.brushes = []


class QuakeMapParser:
    def __init__(self, map_tokenizer):
        self.tokenizer = map_tokenizer

    def _expect_token(self, token_type):
        toktype, tokstring = self.tokenizer.next()
        if toktype != token_type:
            raise ParseError("Expected %s, got %s" %
                    (token.tok_name[token_type],
                        token.tok_name[toktype]))
        return tokstring

    def parse_number(self):
        return float(self._expect_token(token.NUMBER))

    def parse_integer(self):
        return int(self._expect_token(token.NUMBER))

    def parse_quoted_string(self):
        return self._expect_token(token.STRING)[1:-1]

    def expect_newline(self):
        toktype, tokstring = self.tokenizer.next()
        if toktype not in (token.NEWLINE, token.NL):
            raise ParseError("Expected newline, got %s" % tokstring)

    def parse_string(self):
        return self._expect_token(token.NAME)

    def parse_float(self):
        toktype, tokstring = self.tokenizer.next()
        if toktype == token.OP and tokstring == '-':
            return -1.0 * self.parse_number()
        elif toktype == token.NUMBER:
            return float(tokstring)

    def parse_property(self):
        key = self.parse_quoted_string()
        val = self.parse_quoted_string()
        self.expect_newline()
        return key, val

    def expect_specific_string(self, what):
        _, tokstring = self.tokenizer.next()
        if tokstring != what:
            raise ParseError("Expected string %s, got %s" % \
                    (what, tokstring))

    def parse_vertex(self):
        self.expect_specific_string('(')
        x = self.parse_float()
        y = self.parse_float()
        z = self.parse_float()
        self.expect_specific_string(')')
        return x, y, z

    def parse_plane(self):
        plane = Plane()
        plane.x1, plane.y1, plane.z1 = self.parse_vertex()
        plane.x2, plane.y2, plane.z2 = self.parse_vertex()
        plane.x3, plane.y3, plane.z3 = self.parse_vertex()
        plane.texture = self.parse_string()
        plane.xoffset = self.parse_float()
        plane.yoffset = self.parse_float()
        plane.rotation = self.parse_float()
        plane.xscale = self.parse_float()
        plane.yscale = self.parse_float()
        return plane

    def parse_brush(self):
        brush = Brush()
        self.expect_specific_string('{')
        self.expect_newline()
        _, peek_string = self.tokenizer.peek()
        while peek_string == '(':
            brush.planes.append(self.parse_plane())
            self.expect_newline()
            _, peek_string = self.tokenizer.peek()
        self.expect_specific_string('}')
        self.expect_newline()
        return brush

    def parse_entity(self):
        entity = Entity()
        self.expect_specific_string('{')
        self.expect_newline()
        toktype, tokstring = self.tokenizer.peek()
        while toktype == token.STRING:
            k, v = self.parse_property()
            entity.keys[k] = v
            toktype, tokstring = self.tokenizer.peek()
        while toktype == token.NL:
            self.expect_newline()
            toktype, tokstring = self.tokenizer.peek()
        while tokstring == '{':
            entity.brushes.append(self.parse_brush())
            toktype, tokstring = self.tokenizer.peek()
        self.expect_specific_string('}')
        self.expect_newline()
        return entity


if __name__ == '__main__':
    parser = argparse.ArgumentParser(
                        description=__doc__
                        )
    parser.add_argument('source_file',
                        type=argparse.FileType('r', encoding='UTF-8'))
    parsed = parser.parse_args()

    map_tokenizer = IgnoreCommentsTokenizer(parsed.source_file.readline)
    map_parser = QuakeMapParser(map_tokenizer)
    worldspawn = map_parser.parse_entity().brushes
    print('[]Brush {')
    for brush in worldspawn:
        print('    Brush{Planes: []Plane{')
        for plane in brush.planes:
            planestr = '        Plane{X1: %f, Y1: %f, Z1: %f, X2: %f, Y2: %f, Z2: %f, X3: %f, Y3: %f, Z3: %f, Texture: "%s", Xoffset: %f, Yoffset: %f, Rotation: %f, Xscale: %f, Yscale: %f},' % (
				plane.x1, plane.y1, plane.z1,
				plane.x2, plane.y2, plane.z2,
				plane.x3, plane.y3, plane.z3,
				plane.texture,
				plane.xoffset, plane.yoffset,
				plane.rotation,
				plane.xscale, plane.yscale
			)
            print(planestr)
        print('    }},')
    print('}')
