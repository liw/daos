#!/usr/bin/env python3
"""
  (C) Copyright 2024 Intel Corporation.
  (C) Copyright 2025 Hewlett Packard Enterprise Development LP

  SPDX-License-Identifier: BSD-2-Clause-Patent
"""
import ast
import os
import re
import sys
from argparse import ArgumentParser, ArgumentTypeError, RawDescriptionHelpFormatter
from collections import defaultdict
from copy import deepcopy
from pathlib import Path

import yaml

THIS_FILE = os.path.realpath(__file__)
FTEST_DIR = os.path.dirname(THIS_FILE)
MANUAL_TAG = ('manual',)
STAGE_TYPE_TAGS = ('vm', 'hw', 'hw_vmd')
STAGE_SIZE_TAGS = ('medium', 'large')
STAGE_FREQUENCY_TAGS = ('all', 'pr', 'daily_regression', 'full_regression')


class TagSet(set):
    """Set with handling for negative entries."""

    def issubset(self, other):
        for tag in self:
            if tag[0] == '-':
                if tag[1:] in other:
                    return False
            elif tag not in other:
                return False
        return True

    def issuperset(self, other):
        return TagSet.issubset(other, self)


class LintFailure(Exception):
    """Exception for lint failures."""


def all_python_files(path):
    """Get a list of all .py files recursively in a directory.

    Args:
        path (str): directory to look in

    Returns:
        list: sorted path names of .py files
    """
    return sorted(map(str, Path(path).rglob("*.py")))


class FtestTagMap():
    """Represent tags for ftest/avocado."""

    def __init__(self, paths):
        """Initialize the tag mapping.

        Args:
            paths (list): the file or dir path(s) to update from
        """
        self.__mapping = {}  # str(file_name) : str(class_name) : str(test_name) : set(tags)
        for path in paths:
            self.__update_from_path(path)

    def __iter__(self):
        """Iterate over the mapping.

        Yields:
            tuple: file_name, class_name mapping
        """
        for item in self.__mapping.items():
            yield deepcopy(item)

    def __methods(self):
        """Iterate over each method name and its tags.

        Yields:
            (str, set): method name and tags
        """
        for classes in self.__mapping.values():
            for methods in classes.values():
                for method_name, tags in methods.items():
                    yield (method_name, tags)

    def unique_tags(self, exclude=None):
        """Get the set of unique tags, excluding one or more paths.

        Args:
            exclude (list, optional): path(s) to exclude from the unique set.
                Defaults to None.

        Returns:
            set: the set of unique tags
        """
        exclude = list(map(self.__norm_path, exclude or []))

        unique_tags = set()
        for file_path, classes in self.__mapping.items():
            if file_path in exclude:
                continue
            for functions in classes.values():
                for tags in functions.values():
                    unique_tags.update(tags)
        return unique_tags

    def minimal_tags(self, path):
        """Get the minimal tags representing a path in the mapping.

        Args:
            path (str): path to get tags for

        Returns:
            set: set of avocado tag strings

        Raises:
            KeyError: if path is not in the mapping
        """
        file_path = self.__norm_path(path)
        if file_path not in self.__mapping:
            raise KeyError(f'Invalid path: {path}')

        # Keep track of recommended tags for each method
        recommended = set()
        for class_name, functions in self.__mapping[file_path].items():
            for function_name, tags in functions.items():
                # Try the class name and function name first
                if class_name in tags:
                    recommended.add(class_name)
                    continue
                if function_name in tags:
                    recommended.add(function_name)
                    continue
                # Try using a set of tags globally unique to this test.
                # Shouldn't get here since all tests are tagged with their class and method names,
                # but it could happen if the tag lint check is currently failing.
                globally_unique_tags = tags - self.unique_tags(exclude=file_path)
                if globally_unique_tags and globally_unique_tags.issubset(tags):
                    recommended.add(','.join(sorted(globally_unique_tags)))
                    continue
                # Fallback to just using all of this test's tags
                recommended.add(','.join(sorted(tags)))

        return recommended

    def is_test_subset(self, tags1, tags2):
        """Determine whether a set of tags is a subset with respect to tests.

        Args:
            tags1 (list): list of sets of tags
            tags2 (list): list of sets of tags

        Returns:
            bool: whether tags1's tests is a subset of tags2's tests
        """
        tests1 = set(self.__tags_to_tests(tags1))
        tests2 = set(self.__tags_to_tests(tags2))
        return bool(tests1) and bool(tests2) and tests1.issubset(tests2)

    def __tags_to_tests(self, tags):
        """Convert a list of tags to the tests they would run.

        Args:
            tags (list): list of sets of tags
        """
        # Convert to TagSet to handle negative matching
        for idx, _tags in enumerate(tags):
            tags[idx] = TagSet(_tags)
        tests = []
        for method_name, test_tags in self.__methods():
            for tag_set in tags:
                if tag_set.issubset(test_tags):
                    tests.append(method_name)
                    break
        return tests

    def __update_from_path(self, path):
        """Update the mapping from a path.

        Args:
            path (str): the file or dir path to update from

        Raises:
            ValueError: if a path is not a file
        """
        path = self.__norm_path(path)

        if os.path.isdir(path):
            for __path in all_python_files(path):
                self.__parse_file(__path)
            return

        if os.path.isfile(path):
            self.__parse_file(path)
            return

        raise ValueError(f'Expected file or directory: {path}')

    def __parse_file(self, path):
        """Parse a file and update the internal mapping from avocado tags.

        Args:
            path (str): file to parse
        """
        with open(path, 'r') as file:
            file_data = file.read()

        module = ast.parse(file_data)
        for class_def in filter(lambda val: isinstance(val, ast.ClassDef), module.body):
            for func_def in filter(lambda val: isinstance(val, ast.FunctionDef), class_def.body):
                if not func_def.name.startswith('test_'):
                    continue
                tags = self.__parse_avocado_tags(ast.get_docstring(func_def))
                self.__update(path, class_def.name, func_def.name, tags)

    @staticmethod
    def __norm_path(path):
        """Convert to "realpath" and replace .yaml paths with .py equivalents.

        Args:
            path (str): path to normalize

        Returns:
            str: the normalized path
        """
        path = os.path.realpath(path)
        if path.endswith('.yaml'):
            path = re.sub(r'\.yaml$', '.py', path)
        return path

    def __update(self, file_name, class_name, test_name, tags):
        """Update the internal mapping by appending the tags.

        Args:
            file_name (str): file name
            class_name (str): class name
            test_name (str): test name
            tags (set): set of tags to update
        """
        if file_name not in self.__mapping:
            self.__mapping[file_name] = {}
        if class_name not in self.__mapping[file_name]:
            self.__mapping[file_name][class_name] = {}
        if test_name not in self.__mapping[file_name][class_name]:
            self.__mapping[file_name][class_name][test_name] = set()
        self.__mapping[file_name][class_name][test_name].update(tags)

    @staticmethod
    def __parse_avocado_tags(text):
        """Parse avocado tags from a string.

        Args:
            text (str): the string to parse for tags

        Returns:
            set: the set of tags
        """
        tag_strings = re.findall(':avocado: tags=(.*)', text)
        if not tag_strings:
            return set()
        return set(','.join(tag_strings).split(','))


def sorted_tags(tags):
    """Get a sorted list of tags.

    Args:
        tags (set): original tags

    Returns:
        list: sorted tags
    """
    tags_tmp = set(tags)
    new_tags = []
    for tag in STAGE_TYPE_TAGS + STAGE_SIZE_TAGS + STAGE_FREQUENCY_TAGS:
        if tag in tags_tmp:
            new_tags.append(tag)
            tags_tmp.remove(tag)
    new_tags.extend(sorted(tags_tmp))
    return new_tags


def run_linter(paths=None, verbose=False):
    """Run the ftest tag linter.

    Args:
        paths (list, optional): paths to lint. Defaults to all ftest python files

    Raises:
        LintFailure: if linting fails
    """
    if not paths:
        paths = all_python_files(FTEST_DIR)
    all_files = []
    all_classes = defaultdict(int)
    all_methods = defaultdict(int)
    test_wo_tags = []
    tests_wo_class_as_tag = []
    tests_wo_method_as_tag = []
    test_w_invalid_test_tag = []
    tests_wo_hw_vm_manual = []
    tests_w_empty_tag = []
    tests_wo_a_feature_tag = []
    non_feature_tags = set(STAGE_TYPE_TAGS + STAGE_SIZE_TAGS + STAGE_FREQUENCY_TAGS)
    ftest_tag_map = FtestTagMap(paths)
    for file_path, classes in iter(ftest_tag_map):
        all_files.append(file_path)
        for class_name, functions in classes.items():
            all_classes[class_name] += 1
            for method_name, tags in functions.items():
                all_methods[method_name] += 1
                if len(tags) == 0:
                    test_wo_tags.append(method_name)
                if class_name not in tags:
                    tests_wo_class_as_tag.append(method_name)
                if method_name not in tags:
                    tests_wo_method_as_tag.append(method_name)
                for _tag in tags:
                    if _tag.startswith('test_') and _tag != method_name:
                        test_w_invalid_test_tag.append(method_name)
                        break
                if not set(tags).intersection(set(MANUAL_TAG + STAGE_TYPE_TAGS)):
                    tests_wo_hw_vm_manual.append(method_name)
                if '' in tags:
                    tests_w_empty_tag.append(method_name)
                if not set(tags).difference(non_feature_tags | set([class_name, method_name])):
                    tests_wo_a_feature_tag.append(method_name)

    non_unique_classes = list(name for name, num in all_classes.items() if num > 1)
    non_unique_methods = list(name for name, num in all_methods.items() if num > 1)

    def _print_verbose(*args):
        if verbose:
            print(*args)

    _print_verbose('ftest tag lint')

    def _error_handler(_list, message, required=True):
        """Exception handler for each class of failure."""
        _list_len = len(_list)
        req_str = '(required)' if required else '(optional)'
        _print_verbose(f'  {req_str} {_list_len} {message}')
        if _list_len == 0:
            return None
        for _test in _list:
            _print_verbose(f'    {_test}')
        if _list_len > 3:
            remaining = _list_len - 3
            _list = _list[:3] + [f"... (+{remaining})"]
        _list_str = ", ".join(_list)
        if not required:
            # Print but do not fail
            return None
        return LintFailure(f"{_list_len} {message}: {_list_str}")

    # Lint fails if any of the lists contain entries
    errors = list(filter(None, [
        _error_handler(non_unique_classes, 'non-unique test classes'),
        _error_handler(non_unique_methods, 'non-unique test methods'),
        _error_handler(test_wo_tags, 'tests without tags'),
        _error_handler(tests_wo_class_as_tag, 'tests without class as tag'),
        _error_handler(tests_wo_method_as_tag, 'tests without method name as tag'),
        _error_handler(test_w_invalid_test_tag, 'tests with invalid test_ tag'),
        _error_handler(tests_wo_hw_vm_manual, 'tests without HW, VM, or manual tag'),
        _error_handler(tests_w_empty_tag, 'tests with an empty tag'),
        _error_handler(tests_wo_a_feature_tag, 'tests without a feature tag')]))
    if errors:
        raise errors[0]


def run_dump(paths=None, tags=None):
    """Dump the tags per test.

    Formatted as
        <file path>:
          <class name>:
            <method name> - <tags>

    Args:
        paths (list, optional): path(s) to get tags for. Defaults to all ftest python files
        tags2 (list, optional): list of sets of tags to filter.
            Default is None, which does not filter

    Returns:
        int: 0 on success; 1 if no matches found
    """
    if not paths:
        paths = all_python_files(FTEST_DIR)

    # Store output as {path: {class: {test: tags}}}
    output = defaultdict(lambda: defaultdict(dict))

    tag_map = FtestTagMap(paths)
    for file_path, classes in iter(tag_map):
        short_file_path = re.findall(r'ftest/(.*$)', file_path)[0]
        for class_name, functions in classes.items():
            for method_name, method_tags in functions.items():
                if tags and not tag_map.is_test_subset([method_tags], tags):
                    continue
                output[short_file_path][class_name][method_name] = method_tags

    # Format and print output for matching tests
    for short_file_path, classes in output.items():
        print(f'{short_file_path}:')
        for class_name, methods in classes.items():
            print(f'  {class_name}:')
            longest_method_name = max(map(len, methods.keys()))
            for method_name, method_tags in methods.items():
                method_name_fm = method_name.ljust(longest_method_name, " ")
                tags_fm = ",".join(sorted_tags(method_tags))
                print(f'    {method_name_fm} - {tags_fm}')

    return 0 if output else 1


def files_to_tags(paths):
    """Get the unique tags for paths.

    Args:
        paths (list): paths to get tags for

    Returns:
        set: set of test tags representing paths
    """
    # Get tags for ftest paths
    ftest_tag_map = FtestTagMap(all_python_files(FTEST_DIR))

    tag_config = read_tag_config()
    all_tags = set()

    for path in paths:
        for _pattern, _config in tag_config.items():
            if re.search(rf'{_pattern}', path):
                if _config['handler'] == 'FtestTagMap':
                    # Special ftest handling
                    try:
                        all_tags.update(ftest_tag_map.minimal_tags(path))
                    except KeyError:
                        # Use default from config
                        all_tags.update(_config['tags'].split(' '))
                elif _config['handler'] == 'direct':
                    # Use direct tags from config
                    all_tags.update(_config['tags'].split(' '))
                else:
                    raise ValueError(f'Invalid handler: {_config["handler"]}')

                if _config["stop_on_match"]:
                    # Don't process further entries for this path
                    break

    return all_tags


def read_tag_config():
    """Read tags.yaml.

    Also validates the config as read.

    Returns:
        dict: the config

    Raises:
        TypeError: If config types are invalid
        ValueError: If config values are invalid
    """
    with open(os.path.join(FTEST_DIR, 'tags.yaml'), 'r') as file:
        config = yaml.safe_load(file.read())

    for key, val in config.items():
        _type = type(val)
        if _type not in (str, dict):
            raise TypeError(f'Expected {str} or {dict}, not {_type}')
        if _type == str:
            # Expand shorthand to dict
            config[key] = {'tags': val}

        # Check for invalid config options
        extra_keys = set(config[key].keys()) - set(['tags', 'handler', 'stop_on_match'])
        if extra_keys:
            raise ValueError(f'Unsupported keys in config: {", ".join(extra_keys)}')

        # Set default handler and check for invalid values
        if 'handler' not in config[key]:
            config[key]['handler'] = 'direct'
        elif config[key]['handler'] not in ('direct', 'FtestTagMap'):
            raise ValueError(f'Invalid handler: {config[key]["handler"]}')

        # Set default stop_on_match and check for invalid values
        if 'stop_on_match' not in config[key]:
            config[key]['stop_on_match'] = False
        elif not isinstance(config[key]['stop_on_match'], bool):
            raise ValueError(f'Invalid stop_on_match: {config[key]["stop_on_match"]}')

        # Check for missing or invalid tags
        if 'tags' not in config[key]:
            raise ValueError(f'Missing tags for {key}')
        if not isinstance(config[key]['tags'], str):
            raise TypeError(f'Expected {str} for tags, not {type(config[key]["tags"])}')

    return config


def run_list(paths=None):
    """List unique tags for paths.

    Args:
        paths (list, optional): paths to list tags of. Defaults to all ftest python files

    Returns:
        int: 0 on success; 1 if no matches found

    Raises:
        ValueError: if neither paths nor tags is given
    """
    if not paths:
        paths = all_python_files(FTEST_DIR)
    tags = files_to_tags(paths)
    if tags:
        print(' '.join(sorted(tags)))
        return 0
    return 1


def test_tag_set():
    """Run unit tests for TagSet.

    Can be ran directly as:
        tags.py unit
    Or with pytest as:
        python3 -m pytest tags.py

    """
    print('START Ftest TagSet Unit Tests')

    def print_step(*args):
        """Print a step."""
        print('  ', *args)

    l_hw_medium = ['hw', 'medium']
    l_hw_medium_provider = l_hw_medium = ['provider']
    l_hw_medium_minus_provider = l_hw_medium + ['-provider']
    print_step('issubset')
    assert TagSet(l_hw_medium).issubset(l_hw_medium_provider)
    assert not TagSet(l_hw_medium_minus_provider).issubset(l_hw_medium_provider)
    print_step('issuperset')
    assert TagSet(l_hw_medium_provider).issuperset(l_hw_medium)
    assert TagSet(l_hw_medium_provider).issuperset(set(l_hw_medium))
    assert TagSet(l_hw_medium_provider).issuperset(TagSet(l_hw_medium))
    assert not TagSet(l_hw_medium_provider).issuperset(l_hw_medium_minus_provider)
    assert not TagSet(l_hw_medium_provider).issuperset(set(l_hw_medium_minus_provider))
    assert not TagSet(l_hw_medium_provider).issuperset(TagSet(l_hw_medium_minus_provider))
    print('PASS  Ftest TagSet Unit Tests')


def test_tags_util(verbose=False):
    """Run unit tests for FtestTagMap.

    Can be ran directly as:
        tags.py unit
    Or with pytest as:
        python3 -m pytest tags.py

    Args:
        verbose (bool): whether to print verbose output for debugging
    """
    # pylint: disable=protected-access
    print('START Ftest Tags Utility Unit Tests')
    tag_map = FtestTagMap([])
    os.chdir('/')

    def print_step(*args):
        """Print a step."""
        print('  ', *args)

    def print_verbose(*args):
        """Conditionally print if verbose is True."""
        if verbose:
            print('    ', *args)

    print_step('__norm_path')
    assert tag_map._FtestTagMap__norm_path('foo') == '/foo'
    assert tag_map._FtestTagMap__norm_path('/foo') == '/foo'

    print_step('__parse_avocado_tags')
    assert tag_map._FtestTagMap__parse_avocado_tags('') == set()
    assert tag_map._FtestTagMap__parse_avocado_tags('not tags') == set()
    assert tag_map._FtestTagMap__parse_avocado_tags('avocado tags=foo') == set()
    assert tag_map._FtestTagMap__parse_avocado_tags(':avocado: tags=foo') == set(['foo'])
    assert tag_map._FtestTagMap__parse_avocado_tags(':avocado: tags=foo,bar') == set(['foo', 'bar'])
    assert tag_map._FtestTagMap__parse_avocado_tags(
        ':avocado: tags=foo,bar\n:avocado: tags=foo2,bar2') == set(['foo', 'bar', 'foo2', 'bar2'])

    print_step('__update')
    tag_map._FtestTagMap__update('/foo1', 'class_1', 'test_1', set(['class_1', 'test_1', 'foo1']))
    tag_map._FtestTagMap__update('/foo2', 'class_2', 'test_2', set(['class_2', 'test_2', 'foo2']))
    print_verbose(tag_map._FtestTagMap__mapping)
    assert tag_map._FtestTagMap__mapping == {
        '/foo1': {'class_1': {'test_1': {'foo1', 'test_1', 'class_1'}}},
        '/foo2': {'class_2': {'test_2': {'foo2', 'class_2', 'test_2'}}}}

    print_step('__iter__')
    print_verbose(list(iter(tag_map)))
    assert list(iter(tag_map)) == [
        ('/foo1', {'class_1': {'test_1': {'class_1', 'test_1', 'foo1'}}}),
        ('/foo2', {'class_2': {'test_2': {'class_2', 'test_2', 'foo2'}}})
    ]

    print_step('__tags_to_tests')
    assert tag_map._FtestTagMap__tags_to_tests(
        [set(['foo1']), set(['foo2'])]) == ['test_1', 'test_2']
    assert tag_map._FtestTagMap__tags_to_tests([set(['foo1', 'class_1'])]) == ['test_1']

    print_step('__methods')
    assert list(tag_map._FtestTagMap__methods()) == [
        ('test_1', {'class_1', 'test_1', 'foo1'}), ('test_2', {'class_2', 'test_2', 'foo2'})]

    print_step('unique_tags')
    assert tag_map.unique_tags() == set(['class_1', 'test_1', 'foo1', 'class_2', 'test_2', 'foo2'])
    assert tag_map.unique_tags(exclude=['/foo1']) == set(['class_2', 'test_2', 'foo2'])
    assert tag_map.unique_tags(exclude=['/foo2']) == set(['class_1', 'test_1', 'foo1'])
    assert tag_map.unique_tags(exclude=['/foo1', '/foo2']) == set()

    print_step('minimal_tags')
    assert tag_map.minimal_tags('/foo1') == set(['class_1'])
    assert tag_map.minimal_tags('/foo2') == set(['class_2'])

    print_step('is_test_subset')
    assert tag_map.is_test_subset([set(['test_1'])], [set(['test_1'])])
    assert tag_map.is_test_subset([set(['test_1'])], [set(['class_1'])])
    assert tag_map.is_test_subset([set(['test_1', 'foo1'])], [set(['class_1'])])
    assert not tag_map.is_test_subset([set(['test_1'])], [set(['test_2'])])
    assert not tag_map.is_test_subset([set(['test_1'])], [set(['class_2'])])
    assert not tag_map.is_test_subset([set(['test_1', 'foo1'])], [set(['class_2'])])
    assert not tag_map.is_test_subset([set(['test_1']), set(['test_2'])], [set(['class_1'])])
    assert tag_map.is_test_subset([set(['class_2'])], [set(['test_1']), set(['test_2'])])
    assert not tag_map.is_test_subset([set(['fake'])], [set(['class_2'])])

    print_step('__init__')
    # Just a smoke test to verify the map can parse real files
    tag_map = FtestTagMap(all_python_files(FTEST_DIR))
    expected_tags = set(['test_harness_config', 'test_ior_small', 'test_dfuse_mu_perms'])
    assert len(tag_map.unique_tags().intersection(expected_tags)) == len(expected_tags)

    print('PASS  Ftest Tags Utility Unit Tests')


def __arg_type_tags(val):
    """Parse a tags argument.

    Args:
        val (str): string to parse comma-separated tags from

    Returns:
        set: tags converted to a set

    Raises:
        ArgumentTypeError: if val is invalid
    """
    if not val:
        raise ArgumentTypeError("tags cannot be empty")
    try:
        return set(map(str.strip, val.split(",")))
    except Exception as err:  # pylint: disable=broad-except
        raise ArgumentTypeError(f"Invalid tags: {val}") from err


def main():
    """main function execution"""
    description = '\n'.join([
        'Commands',
        '  lint - lint ftest avocado tags',
        '  list - list ftest avocado tags associated with test files',
        '  dump - dump the file/class/method/tag structure for test files',
        '  unit - run self unit tests'
    ])

    parser = ArgumentParser(formatter_class=RawDescriptionHelpFormatter, description=description)
    parser.add_argument(
        "command",
        choices=("lint", "list", "dump", "unit"),
        help="command to run")
    parser.add_argument(
        "-v", "--verbose",
        action='store_true',
        help="print verbose output for some commands")
    parser.add_argument(
        "--paths",
        nargs="+",
        default=[],
        help="file paths")
    parser.add_argument(
        "--tags",
        nargs="+",
        type=__arg_type_tags,
        help="tags")
    args = parser.parse_args()
    args.paths = list(map(os.path.realpath, args.paths))

    # Check for incompatible arguments
    rc = 0
    if args.command == "lint" and args.tags:
        print("--tags not supported with lint")
        rc = 1
    if args.command == "list" and args.tags:
        print("--tags not supported with list")
        rc = 1
    if args.command == "unit" and args.tags:
        print("--tags not supported with unit")
        rc = 1
    if args.command == "unit" and args.paths:
        print("--paths not supported with unit")
        rc = 1
    if rc != 0:
        return rc

    if args.command == "lint":
        try:
            run_linter(args.paths, args.verbose)
            rc = 0
        except LintFailure as err:
            print(err)
            rc = 1
    elif args.command == "dump":
        rc = run_dump(args.paths, args.tags)
    elif args.command == "list":
        rc = run_list(args.paths)
    elif args.command == "unit":
        test_tag_set()
        test_tags_util(args.verbose)
        rc = 0

    return rc


if __name__ == '__main__':
    sys.exit(main())
