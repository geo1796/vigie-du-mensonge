import {useMemo, useState} from "react";
import type {User} from "@/core/models/user.ts";
import type {AdminClient} from "@/core/dependencies/admin/adminClient.ts";
import {type UserRole, UserRoleLabels, UserRoles} from "@/core/models/userRole.ts";
import {Popover, PopoverContent, PopoverTrigger} from "@/core/shadcn/components/ui/popover.tsx";
import {Plus, X} from "lucide-react";
import {toast} from "@/core/utils/toast.ts";
import {fmtDate} from "@/core/utils/fmtDate.ts";

export type AdminUserProfileProps = {
    user: User;
    adminClient: AdminClient;
}

export function AdminUserProfile({user, adminClient}: AdminUserProfileProps) {
    const [roles, setRoles] = useState<UserRole[]>(user.roles ?? []);
    const [addOpen, setAddOpen] = useState(false);

    const allRoles = useMemo(() => Object.keys(UserRoles) as UserRole[], []);
    const missingRoles = useMemo(() => allRoles.filter(r => !roles.includes(r)), [allRoles, roles]);

    const handleRevoke = async (role: UserRole) => {
        // Prevent actions on users who are ADMIN
        if (roles.includes(UserRoles.ADMIN as UserRole)) {
            toast.error("Cet utilisateur est administrateur. Action non autorisée.");
            return;
        }
        try {
            await adminClient.revokeUserRole(user.tag, role);
            setRoles(prev => prev.filter(r => r !== role));
            toast.success("Rôle retiré.");
        } catch (e) {
            console.error(e);
            toast.error("Une erreur est survenue. Veuillez réessayer.");
        }
    };

    const handleGrant = async (role: UserRole) => {
        try {
            await adminClient.grantUserRole(user.tag, role);
            setRoles(prev => prev.includes(role) ? prev : [...prev, role]);
            toast.success("Rôle ajouté.");
            setAddOpen(false);
        } catch (e) {
            console.error(e);
            toast.error("Une erreur est survenue. Veuillez réessayer.");
        }
    };

    return (
        <div className="flex flex-col items-center gap-6">
            <div className="space-y-2">
                <h2 className="text-xl font-semibold">Profil utilisateur</h2>
                <div className="mt-2 text-sm text-muted-foreground">Tag: <span
                    className="font-medium text-foreground">{user.tag}</span></div>
                <div className="text-sm text-muted-foreground">Créé le: <span
                    className="font-medium text-foreground">{fmtDate(user.createdAt)}</span></div>
            </div>

            <div className="flex items-center justify-center gap-2">
                <h3 className="text-lg font-medium">Rôles</h3>
                <Popover open={addOpen} onOpenChange={setAddOpen}>
                    <PopoverTrigger asChild>
                        <button
                            type="button"
                            aria-label="Ajouter un rôle"
                            className="inline-flex items-center gap-2 rounded-md border px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground"
                        >
                            <Plus className="h-4 w-4"/>
                            Ajouter
                        </button>
                    </PopoverTrigger>
                    <PopoverContent className="w-56 p-2">
                        <div className="text-sm font-medium mb-2">Ajouter un rôle</div>
                        {missingRoles.length === 0 ? (
                            <div className="text-sm text-muted-foreground">Aucun rôle disponible</div>
                        ) : (
                            <ul className="space-y-1">
                                {missingRoles.map((r) => (
                                    <li key={r}>
                                        <button
                                            type="button"
                                            onClick={() => handleGrant(r)}
                                            className="w-full text-left rounded-md px-2 py-1.5 hover:bg-accent hover:text-accent-foreground"
                                        >
                                            {UserRoleLabels[r]}
                                        </button>
                                    </li>
                                ))}
                            </ul>
                        )}
                    </PopoverContent>
                </Popover>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 gap-3">
                {roles.map((role) => (
                    <div key={role} className="relative rounded-md border p-4">
                        <button
                            type="button"
                            aria-label={`Retirer le rôle ${UserRoleLabels[role]}`}
                            onClick={() => handleRevoke(role)}
                            className="absolute right-2 top-2 rounded-full p-1 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                        >
                            <X className="h-4 w-4"/>
                        </button>
                        <div className="text-sm text-muted-foreground">Rôle</div>
                        <div className="text-base font-medium">{UserRoleLabels[role]}</div>
                        <div className="mt-1 text-xs text-muted-foreground">({role})</div>
                    </div>
                ))}
                {roles.length === 0 && (
                    <div className="text-sm text-muted-foreground">Cet utilisateur n'a aucun rôle.</div>
                )}
            </div>
        </div>
    );
}