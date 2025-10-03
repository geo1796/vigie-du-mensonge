import {useEffect, useState} from "react";
import {useQuery} from "@tanstack/react-query";
import {Input} from "@/core/shadcn/components/ui/input.tsx";
import {Spinner} from "@/core/shadcn/components/ui/spinner.tsx";
import type {AdminClient} from "@/core/dependencies/admin/adminClient.ts";
import {Link} from "@/core/utils/router.ts";

export type AdminUsersSearchBarProps = {
    adminClient: AdminClient;
}

export function AdminUsersSearchBar({adminClient}: AdminUsersSearchBarProps) {
    const [term, setTerm] = useState<string>("");
    const [debounced, setDebounced] = useState<string>("");

    useEffect(() => {
        const handler = setTimeout(() => setDebounced(term.trim()), 300);
        return () => clearTimeout(handler);
    }, [term]);

    const enabled = debounced.length >= 2;

    const {queryKey, queryFn} = adminClient.searchUsersByTag(debounced);

    const {data: results, isFetching, isError} = useQuery({
        queryKey,
        queryFn,
        enabled,
        staleTime: 5 * 60 * 1000, //5 minutes
        gcTime: 10 * 60 * 1000, // 10 minutes
    });

    return (
        <div className="flex flex-col gap-2 w-full max-w-xl">
            <Input
                value={term}
                onChange={(e) => setTerm(e.target.value)}
                placeholder="Rechercher un utilisateur par tag"
            />

            {!enabled && term.length > 0 && (
                <div className="text-xs text-muted-foreground">Entrez au moins 2 caractères pour rechercher.</div>
            )}

            {enabled && (
                <div className="border rounded-md p-2">
                    {isFetching && (
                        <div className="flex items-center gap-2 text-sm text-muted-foreground">
                            <Spinner size={16}/> Recherche en cours...
                        </div>
                    )}

                    {isError && (
                        <div className="text-sm text-destructive">Une erreur est survenue. Veuillez réessayer.</div>
                    )}

                    {results && results.length === 0 && !isFetching && !isError && (
                        <div className="text-sm text-muted-foreground">Aucun résultat</div>
                    )}

                    {results && results.length > 0 && (
                        <ul className="flex flex-col gap-1">
                            {results.map((tag) => (
                                <Link to={`/admin/users/$userTag`} params={{userTag: tag}}
                                      key={tag} className="text-sm px-2 py-1 rounded hover:bg-accent">
                                    {tag}
                                </Link>
                            ))}
                        </ul>
                    )}
                </div>
            )}
        </div>
    );
}