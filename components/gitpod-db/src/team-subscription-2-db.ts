/**
 * Copyright (c) 2021 Gitpod GmbH. All rights reserved.
 * Licensed under the Gitpod Enterprise Source Code License,
 * See License.enterprise.txt in the project root folder.
 */

import { TeamSubscription2 } from "@gitpod/gitpod-protocol/lib/team-subscription-protocol";

export const TeamSubscription2DB = Symbol('TeamSubscription2DB');
export interface TeamSubscription2DB {
    storeTeamSubscription2Entry(ts: TeamSubscription2): Promise<void>;
    findTeamSubscription2ById(id: string): Promise<TeamSubscription2 | undefined>;
    findTeamSubscription2ByPaymentRef(teamId: string, paymentReference: string): Promise<TeamSubscription2 | undefined>;

    transaction<T>(code: (db: TeamSubscription2DB) => Promise<T>): Promise<T>;
}