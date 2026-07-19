import { FormEvent, useState } from "react";
import { formatDistanceToNow } from "date-fns";
import { Pencil, Send, Trash2, X } from "lucide-react";
import { toast } from "sonner";

import type { Comment } from "@/api/types";
import { getApiErrorMessage } from "@/api/http";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Textarea } from "@/components/ui/textarea";
import {
  useCommentsForPhoto,
  useCreateComment,
  useDeleteComment,
  useUpdateComment,
} from "@/hooks/use-comments";
import { ownerId } from "@/lib/utils";
import { useAuthStore } from "@/stores/auth-store";

export function CommentList({ photoId }: { photoId: number }) {
  const [message, setMessage] = useState("");
  const comments = useCommentsForPhoto(photoId);
  const createComment = useCreateComment(photoId);

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    try {
      await createComment.mutateAsync({ message });
      setMessage("");
      toast.success("Comment posted");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle>Comments</CardTitle>
      </CardHeader>
      <CardContent className="space-y-5">
        <form onSubmit={handleSubmit} className="flex flex-col gap-3 sm:flex-row">
          <Textarea
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            placeholder="Write a comment"
            required
            className="min-h-16"
          />
          <Button type="submit" disabled={createComment.isPending} className="sm:self-start">
            <Send className="mr-2 h-4 w-4" aria-hidden="true" />
            Send
          </Button>
        </form>

        <div className="space-y-3">
          {comments.data?.map((comment) => (
            <CommentItem key={comment.id} photoId={photoId} comment={comment} />
          ))}

          {!comments.isLoading && comments.data?.length === 0 ? (
            <p className="text-sm text-muted-foreground">No comments yet.</p>
          ) : null}
        </div>
      </CardContent>
    </Card>
  );
}

function CommentItem({ photoId, comment }: { photoId: number; comment: Comment }) {
  const [editing, setEditing] = useState(false);
  const [message, setMessage] = useState(comment.message);
  const currentUser = useAuthStore((state) => state.user);
  const updateComment = useUpdateComment(photoId);
  const deleteComment = useDeleteComment(photoId);
  const isOwner = currentUser?.id === ownerId(comment);
  const createdAt = comment.created_at
    ? formatDistanceToNow(new Date(comment.created_at), { addSuffix: true })
    : null;

  async function handleUpdate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();

    try {
      await updateComment.mutateAsync({ commentId: comment.id, payload: { message } });
      setEditing(false);
      toast.success("Comment updated");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  async function handleDelete() {
    if (!window.confirm("Delete this comment?")) {
      return;
    }

    try {
      await deleteComment.mutateAsync(comment.id);
      toast.success("Comment deleted");
    } catch (error) {
      toast.error(getApiErrorMessage(error));
    }
  }

  return (
    <div className="rounded-md border bg-background p-3">
      {editing ? (
        <form onSubmit={handleUpdate} className="grid gap-2">
          <Textarea
            value={message}
            onChange={(event) => setMessage(event.target.value)}
            aria-label="Comment message"
            required
          />
          <div className="flex flex-wrap gap-2">
            <Button type="submit" size="sm" disabled={updateComment.isPending}>
              Save
            </Button>
            <Button type="button" size="sm" variant="outline" onClick={() => setEditing(false)}>
              <X className="mr-2 h-4 w-4" aria-hidden="true" />
              Cancel
            </Button>
          </div>
        </form>
      ) : (
        <p className="text-sm">{comment.message}</p>
      )}
      <div className="mt-2 flex items-center justify-between gap-3">
        {createdAt ? <p className="text-xs text-muted-foreground">{createdAt}</p> : <span />}
        {isOwner ? (
          <div className="flex gap-1">
            <Button
              type="button"
              size="icon"
              variant="ghost"
              onClick={() => setEditing((current) => !current)}
              aria-label="Edit comment"
            >
              <Pencil className="h-4 w-4" aria-hidden="true" />
            </Button>
            <Button
              type="button"
              size="icon"
              variant="ghost"
              onClick={handleDelete}
              aria-label="Delete comment"
            >
              <Trash2 className="h-4 w-4" aria-hidden="true" />
            </Button>
          </div>
        ) : null}
      </div>
    </div>
  );
}
