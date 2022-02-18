/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the GNU Affero General Public License (AGPL).
 * See License-AGPL.txt in the project root for license information.
 */

export function timeout(millis: number): AbortController {
    const controller = new AbortController();
    setTimeout(controller.abort, millis);
    return controller;
}