# -*- Mode:Python; indent-tabs-mode:nil; tab-width:4 -*-
#
# Copyright (C) 2016-2017 Canonical Ltd
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License version 3 as
# published by the Free Software Foundation.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

"""Bower plugin just run bower install after copying whole source content.

Copy is performed by inheriting from the dump plugin.
"""

import os
import shutil

import snapcraft
from snapcraft.plugins import dump, nodejs


class BowerPlugin(dump.DumpPlugin, nodejs.NodePlugin):

    def __init__(self, name, options, project):
        super().__init__(name, options, project)
        options.node_packages += ["bower"]

    def build(self):
        ''''Setup build and install directory with source sets'''
        super().build()

        # remove locally built bower components
        shutil.rmtree(os.path.join(self.installdir, 'bower_components'), ignore_errors=True)

        self.run(['bower', '--allow-root', 'install'], cwd=self.installdir)
