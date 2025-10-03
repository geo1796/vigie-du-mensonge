import {createFileRoute} from '@tanstack/react-router';
import {AdminUsersSearchBar} from "@/core/components/admin/AdminUsersSearchBar.tsx";

export const Route = createFileRoute('/admin/users/')({
    component: RouteComponent,
});

function RouteComponent() {
    const adminClient = Route.useRouteContext().adminClient;

    return <div className="flex flex-col items-center gap-8 min-w-0 p-2">
        <AdminUsersSearchBar adminClient={adminClient}/>
    </div>;
}
