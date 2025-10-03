import {Button} from "@/core/shadcn/components/ui/button.tsx";
import {Sheet, SheetClose, SheetContent, SheetTitle, SheetTrigger} from "@/core/shadcn/components/ui/sheet.tsx";
import {useStore} from "@tanstack/react-store";
import {authManager} from "@/core/dependencies/auth/authManager.ts";
import {Link} from "@/core/utils/router.ts";
import {Menu} from "lucide-react";
import {ThemeToggle} from "@/core/components/misc/ThemeToggle.tsx";
import {Separator} from "@/core/shadcn/components/ui/separator.tsx";

export function SideSheet() {
    const auth = useStore(authManager.authStore);

    return <Sheet>
        <SheetTrigger asChild>
            <Button variant="ghost"><Menu/></Button>
        </SheetTrigger>
        <SheetContent aria-describedby={undefined} side="right" className="w-64">
            <div className="flex flex-col items-center gap-4 mt-8 mx-4">

                <ThemeToggle/>

                <SheetTitle className="p-4">Navigation</SheetTitle>
                <Separator/>

                {!auth ?
                    <>
                        <SheetClose asChild>
                            <Link
                                to="/sign-in"
                                className="p-2 text-sm font-medium rounded-md hover:bg-accent"
                            >
                                Connexion
                            </Link>
                        </SheetClose>
                        <SheetClose asChild>
                            <Link
                                to="/sign-up" search={{token: undefined}}
                                className="p-2 text-sm font-medium rounded-md hover:bg-accent"
                            >
                                Inscription
                            </Link>
                        </SheetClose>
                    </>
                    :
                    <>
                        {auth.isAdmin &&
                            <SheetClose asChild>
                                <Link
                                    to="/admin/users"
                                    className="p-2 text-sm font-medium rounded-md hover:bg-accent"
                                >
                                    Espace administrateur
                                </Link>
                            </SheetClose>
                        }
                        {auth.isModerator &&
                            <SheetClose asChild>
                                <Link
                                    to="/moderator/articles"
                                    className="p-2 text-sm font-medium rounded-md hover:bg-accent"
                                >
                                    Espace modérateur
                                </Link>
                            </SheetClose>
                        }
                        {auth.isRedactor &&
                            <SheetClose asChild>
                                <Link
                                    to="/redactor/articles"
                                    className="p-2 text-sm font-medium rounded-md hover:bg-accent"
                                >
                                    Espace rédacteur
                                </Link>
                            </SheetClose>
                        }
                        <SheetClose asChild>
                            <Button
                                onClick={() => authManager.signOut()}>Déconnexion</Button>
                        </SheetClose>
                    </>
                }
            </div>
        </SheetContent>
    </Sheet>;
}