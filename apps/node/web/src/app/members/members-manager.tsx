"use client";

import { FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useToast } from "@/components/ui/toast-provider";
import { readApiErrorMessage, toErrorMessage } from "@/lib/api/error";
import type { components } from "@/lib/api/generated";

type TenantMember = components["schemas"]["TenantMember"];
type TenantRole = components["schemas"]["TenantRole"];

export function MembersManager({
  initialMembers,
  currentUserId,
}: {
  initialMembers: TenantMember[];
  currentUserId: string;
}) {
  const { pushToast } = useToast();
  const [members, setMembers] = useState(initialMembers);
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<TenantRole>("member");
  const [loading, setLoading] = useState(false);

  const currentMember = members.find((m) => m.user_id === currentUserId);
  const canManage =
    currentMember?.role === "owner" || currentMember?.role === "admin";

  async function refreshMembers() {
    const res = await fetch("/api/members", { cache: "no-store" });
    if (!res.ok) {
      const message = await readApiErrorMessage(res, "Failed to load members");
      throw new Error(message);
    }
    const data = (await res.json()) as { items: TenantMember[] };
    setMembers(data.items ?? []);
  }

  async function handleInvite(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setLoading(true);
    try {
      const res = await fetch("/api/invitations", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ email: inviteEmail, role: inviteRole }),
      });
      if (!res.ok) {
        const message = await readApiErrorMessage(res, "Failed to invite");
        throw new Error(message);
      }
      await refreshMembers();
      setInviteEmail("");
      setInviteRole("member");
      pushToast({ variant: "success", message: "Invitation sent" });
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to invite"),
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleRoleChange(userId: string, role: TenantRole) {
    setLoading(true);
    try {
      const res = await fetch(`/api/members/${userId}`, {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ role }),
      });
      if (!res.ok) {
        const message = await readApiErrorMessage(
          res,
          "Failed to update role"
        );
        throw new Error(message);
      }
      await refreshMembers();
      pushToast({ variant: "success", message: "Role updated" });
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to update role"),
      });
    } finally {
      setLoading(false);
    }
  }

  async function handleRemove(userId: string) {
    const isSelf = userId === currentUserId;
    const confirmed = window.confirm(
      isSelf
        ? "Are you sure you want to leave this tenant?"
        : "Are you sure you want to remove this member?"
    );
    if (!confirmed) return;

    setLoading(true);
    try {
      const res = await fetch(`/api/members/${userId}`, {
        method: "DELETE",
      });
      if (!res.ok && res.status !== 204) {
        const message = await readApiErrorMessage(
          res,
          "Failed to remove member"
        );
        throw new Error(message);
      }
      if (isSelf) {
        window.location.href = "/dashboard";
        return;
      }
      await refreshMembers();
      pushToast({ variant: "success", message: "Member removed" });
    } catch (error) {
      pushToast({
        variant: "error",
        message: toErrorMessage(error, "Failed to remove member"),
      });
    } finally {
      setLoading(false);
    }
  }

  const roles: TenantRole[] = ["owner", "admin", "member"];

  return (
    <div className="space-y-8">
      {canManage && (
        <form
          onSubmit={handleInvite}
          className="space-y-4 rounded-lg border p-4"
        >
          <h2 className="text-lg font-semibold">Invite Member</h2>
          <div className="grid gap-4 md:grid-cols-3">
            <div className="space-y-2">
              <Label htmlFor="invite-email">Email</Label>
              <Input
                id="invite-email"
                type="email"
                value={inviteEmail}
                onChange={(e) => setInviteEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="invite-role">Role</Label>
              <select
                id="invite-role"
                className="h-9 w-full rounded-md border bg-background px-3 text-sm"
                value={inviteRole}
                onChange={(e) => setInviteRole(e.target.value as TenantRole)}
              >
                <option value="member">member</option>
                <option value="admin">admin</option>
                {currentMember?.role === "owner" && (
                  <option value="owner">owner</option>
                )}
              </select>
            </div>
            <div className="flex items-end">
              <Button type="submit" disabled={loading}>
                Invite
              </Button>
            </div>
          </div>
        </form>
      )}

      <div className="rounded-lg border">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-muted/50">
              <th className="px-4 py-3 text-left font-medium">Email</th>
              <th className="px-4 py-3 text-left font-medium">Display Name</th>
              <th className="px-4 py-3 text-left font-medium">Role</th>
              <th className="px-4 py-3 text-left font-medium">Joined</th>
              <th className="px-4 py-3 text-left font-medium">Actions</th>
            </tr>
          </thead>
          <tbody>
            {members.map((member) => {
              const isSelf = member.user_id === currentUserId;
              const canChangeRole =
                canManage &&
                !isSelf &&
                !(
                  currentMember?.role === "admin" &&
                  (member.role === "owner" || false)
                );
              const canRemove =
                (canManage &&
                  !isSelf &&
                  !(
                    currentMember?.role === "admin" &&
                    member.role === "owner"
                  )) ||
                isSelf;

              return (
                <tr
                  key={member.user_id}
                  className={`border-b last:border-0 ${isSelf ? "bg-muted/30" : ""}`}
                >
                  <td className="px-4 py-3">
                    {member.email}
                    {isSelf && (
                      <span className="ml-2 text-xs text-muted-foreground">
                        (you)
                      </span>
                    )}
                  </td>
                  <td className="px-4 py-3">{member.display_name}</td>
                  <td className="px-4 py-3">
                    {canChangeRole ? (
                      <select
                        className="h-8 rounded-md border bg-background px-2 text-sm"
                        value={member.role}
                        onChange={(e) =>
                          handleRoleChange(
                            member.user_id,
                            e.target.value as TenantRole
                          )
                        }
                        disabled={loading}
                      >
                        {roles.map((r) => (
                          <option
                            key={r}
                            value={r}
                            disabled={
                              currentMember?.role === "admin" && r === "owner"
                            }
                          >
                            {r}
                          </option>
                        ))}
                      </select>
                    ) : (
                      member.role
                    )}
                  </td>
                  <td className="px-4 py-3 text-muted-foreground">
                    {new Date(member.joined_at).toLocaleDateString()}
                  </td>
                  <td className="px-4 py-3">
                    {canRemove && (
                      <Button
                        size="sm"
                        variant="destructive"
                        onClick={() => handleRemove(member.user_id)}
                        disabled={loading}
                      >
                        {isSelf ? "Leave" : "Remove"}
                      </Button>
                    )}
                  </td>
                </tr>
              );
            })}
            {members.length === 0 && (
              <tr>
                <td
                  colSpan={5}
                  className="px-4 py-6 text-center text-muted-foreground"
                >
                  No members yet.
                </td>
              </tr>
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}
