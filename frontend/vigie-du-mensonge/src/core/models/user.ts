import type {UserRole} from "@/core/models/userRole.ts";

export type UserJson = {
    tag: string;
    createdAt: string;
    roles: UserRole[];
}

export class User {
    public tag: string;
    public createdAt: Date;
    public roles: UserRole[];

    constructor(tag: string, createdAt: Date, roles: UserRole[]) {
        this.tag = tag;
        this.createdAt = createdAt;
        this.roles = roles;
    }

    static fromJson(json: UserJson): User {
        return new User(json.tag, new Date(json.createdAt), json.roles);
    }
}