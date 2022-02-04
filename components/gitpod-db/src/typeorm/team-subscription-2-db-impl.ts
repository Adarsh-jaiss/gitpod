/**
 * Copyright (c) 2022 Gitpod GmbH. All rights reserved.
 * Licensed under the Gitpod Enterprise Source Code License,
 * See License.enterprise.txt in the project root folder.
 */

import { injectable, inject } from "inversify";
import { EntityManager, Repository } from "typeorm";

import { TeamSubscription2 } from "@gitpod/gitpod-protocol/lib/team-subscription-protocol";

import { TeamSubscription2DB } from "../team-subscription-2-db";
import { DBTeamSubscription2 } from "./entity/db-team-subscription-2";
import { TypeORM } from "./typeorm";

@injectable()
export class TeamSubscription2DBImpl implements TeamSubscription2DB {
    @inject(TypeORM) protected readonly typeORM: TypeORM;

    async transaction<T>(code: (db: TeamSubscription2DB) => Promise<T>): Promise<T> {
        const manager = await this.getEntityManager();
        return await manager.transaction(async manager => {
            return await code(new TransactionalTeamSubscription2DBImpl(manager));
        });
    }

    protected async getEntityManager() {
        return (await this.typeORM.getConnection()).manager;
    }

    protected async getRepo(): Promise<Repository<DBTeamSubscription2>> {
        return (await this.getEntityManager()).getRepository(DBTeamSubscription2);
    }

    /**
     * Team Subscriptions 2
     */

    async storeTeamSubscription2Entry(ts: TeamSubscription2): Promise<void> {
        const repo = await this.getRepo();
        await repo.save(ts);
    }

    async findTeamSubscription2ById(id: string): Promise<TeamSubscription2 | undefined> {
        const repo = await this.getRepo();
        return repo.findOne(id);
    }

    async findTeamSubscription2ByPaymentRef(teamId: string, paymentReference: string): Promise<TeamSubscription2 | undefined> {
        const repo = await this.getRepo();
        return repo.findOne({ teamId, paymentReference });
    }

}

export class TransactionalTeamSubscription2DBImpl extends TeamSubscription2DBImpl {
    constructor(protected readonly manager: EntityManager) {
        super();
    }

    async getEntityManager(): Promise<EntityManager> {
        return this.manager;
    }
}