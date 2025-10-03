import type {KyInstance} from "ky";
import {User, type UserJson} from "@/core/models/user.ts";

export class AdminClient {
    private readonly api: KyInstance;

    constructor(api: KyInstance) {
        this.api = api;
    }

    private async _searchUsersByTag(tag: string): Promise<string[]> {
        const res = await this.api
            .get(`admin/users?userTag=${tag}`)
            .json<{ results: string[] }>();

        return res.results;
    }

    searchUsersByTag = (tag: string): { queryKey: string[], queryFn: () => Promise<string[]> } => {
        return {
            queryKey: ["admin", "searchUsers", tag],
            queryFn: () => this._searchUsersByTag(tag),
        };
    };

    private async _findUserByTag(tag: string): Promise<User> {
        const res = await this.api
            .get(`admin/users/${tag}`)
            .json<UserJson>();

        return User.fromJson(res);
    }

    findUserByTag = (tag: string): { queryKey: string[], queryFn: () => Promise<User> } => {
        return {
            queryKey: ["admin", "users", tag],
            queryFn: () => this._findUserByTag(tag),
        };
    };

    async grantUserRole(userTag: string, roleName: string): Promise<void> {
        await this.api.post(`admin/users/${userTag}/roles/${roleName}`);
    }

    async revokeUserRole(userTag: string, roleName: string): Promise<void> {
        await this.api.delete(`admin/users/${userTag}/roles/${roleName}`);
    }
}