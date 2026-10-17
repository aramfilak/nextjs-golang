"use client";

import { useEffect, useRef, useState, type FormEvent } from "react";
import styles from "./page.module.css";

type Todo = { id: string; title: string; completed: boolean };
type Filter = "all" | "open" | "completed";

const filters: { value: Filter; label: string }[] = [
  { value: "all", label: "Alle" },
  { value: "open", label: "Offen" },
  { value: "completed", label: "Erledigt" },
];

async function request<T>(path = "", options?: RequestInit): Promise<T> {
  const response = await fetch(`/api/todos${path}`, {
    ...options,
    headers: { "Content-Type": "application/json", ...options?.headers },
    cache: "no-store",
  });
  if (!response.ok)
    throw new Error("Die Anfrage konnte nicht gespeichert werden.");
  if (response.status === 204) return undefined as T;
  return response.json();
}

export default function Home() {
  const [todos, setTodos] = useState<Todo[]>([]);
  const [title, setTitle] = useState("");
  const [filter, setFilter] = useState<Filter>("all");
  const [loading, setLoading] = useState(true);
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState("");
  const inputRef = useRef<HTMLInputElement>(null);

  async function load(signal?: AbortSignal) {
    try {
      const items = await request<Todo[]>("", { signal });
      setTodos(items);
      setError("");
    } catch {
      if (!signal?.aborted)
        setError(
          "Aufgaben konnten nicht geladen werden. Bitte versuche es erneut.",
        );
    } finally {
      if (!signal?.aborted) setLoading(false);
    }
  }

  useEffect(() => {
    const controller = new AbortController();
    void load(controller.signal);
    return () => controller.abort();
  }, []);

  async function mutate(action: () => Promise<void>) {
    setBusy(true);
    setError("");
    try {
      await action();
    } catch {
      setError("Änderung fehlgeschlagen. Bitte versuche es erneut.");
    } finally {
      setBusy(false);
    }
  }

  async function addTodo(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!title.trim() || busy) return;
    await mutate(async () => {
      const item = await request<Todo>("", {
        method: "POST",
        body: JSON.stringify({ title: title.trim() }),
      });
      setTodos((items) => [...items, item]);
      setTitle("");
    });
    inputRef.current?.focus();
  }

  function toggleTodo(todo: Todo) {
    void mutate(async () => {
      const updated = await request<Todo>(`/${todo.id}`, {
        method: "PATCH",
        body: JSON.stringify({ completed: !todo.completed }),
      });
      setTodos((items) =>
        items.map((item) => (item.id === updated.id ? updated : item)),
      );
    });
  }

  function deleteTodo(id: string) {
    void mutate(async () => {
      await request<void>(`/${id}`, { method: "DELETE" });
      setTodos((items) => items.filter((item) => item.id !== id));
    });
  }

  const completed = todos.filter((todo) => todo.completed).length;
  const visibleTodos = todos.filter(
    (todo) => filter === "all" || todo.completed === (filter === "completed"),
  );

  return (
    <main className={styles.page}>
      <div className={styles.container}>
        <header className={styles.header}>
          <span className={styles.eyebrow}>EIN SCHRITT NACH DEM ANDEREN</span>
          <h1>
            Meine Aufgaben<span>.</span>
          </h1>
          <p>Weniger im Kopf. Mehr geschafft.</p>
        </header>

        <section
          className={styles.card}
          aria-label="Todo-Liste"
          aria-busy={loading || busy}
        >
          <form className={styles.form} onSubmit={addTodo}>
            <label htmlFor="todo-title">Was möchtest du erledigen?</label>
            <div className={styles.inputRow}>
              <input
                ref={inputRef}
                id="todo-title"
                value={title}
                onChange={(event) => setTitle(event.target.value)}
                maxLength={200}
                required
                autoComplete="off"
                disabled={loading || busy}
              />
              <button
                className={styles.add}
                type="submit"
                disabled={loading || busy || !title.trim()}
              >
                <span aria-hidden="true">＋</span> Hinzufügen
              </button>
            </div>
          </form>

          {error && (
            <div className={styles.error} role="alert">
              {error}{" "}
              <button
                type="button"
                disabled={loading || busy}
                onClick={() => {
                  setLoading(true);
                  void load();
                }}
              >
                Neu laden
              </button>
            </div>
          )}

          <div className={styles.toolbar}>
            <div
              className={styles.filters}
              role="group"
              aria-label="Aufgaben filtern"
            >
              {filters.map((item) => (
                <button
                  key={item.value}
                  type="button"
                  aria-pressed={filter === item.value}
                  onClick={() => setFilter(item.value)}
                >
                  {item.label}
                </button>
              ))}
            </div>
            <span className={styles.count} aria-live="polite">
              {todos.length - completed} offen
            </span>
          </div>

          {loading ? (
            <p className={styles.empty} role="status">
              Aufgaben werden geladen …
            </p>
          ) : visibleTodos.length === 0 ? (
            <div className={styles.empty}>
              <span className={styles.emptyIcon} aria-hidden="true">
                ✓
              </span>
              <h2>
                {todos.length === 0
                  ? "Platz für deinen nächsten Schritt"
                  : "Alles im Blick"}
              </h2>
              <p>
                {todos.length === 0
                  ? "Trage oben deine erste Aufgabe ein."
                  : filter === "open"
                    ? "Alle Aufgaben sind erledigt. Gut gemacht!"
                    : "Hier gibt es gerade keine Aufgaben."}
              </p>
            </div>
          ) : (
            <ul className={styles.list}>
              {visibleTodos.map((todo) => (
                <li
                  key={todo.id}
                  className={todo.completed ? styles.completed : undefined}
                >
                  <label className={styles.todo}>
                    <input
                      type="checkbox"
                      checked={todo.completed}
                      disabled={busy}
                      onChange={() => toggleTodo(todo)}
                    />
                    <span>{todo.title}</span>
                  </label>
                  <button
                    className={styles.delete}
                    type="button"
                    disabled={busy}
                    onClick={() => deleteTodo(todo.id)}
                    aria-label={`Aufgabe „${todo.title}“ löschen`}
                  >
                    <svg
                      width="18"
                      height="18"
                      viewBox="0 0 24 24"
                      fill="none"
                      stroke="currentColor"
                      strokeWidth="1.6"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      aria-hidden="true"
                    >
                      <path d="M3 6h18M9 6V4h6v2M5 6l1 14h12l1-14M10 10v6M14 10v6" />
                    </svg>
                  </button>
                </li>
              ))}
            </ul>
          )}

          <footer className={styles.progress}>
            <span>
              {completed} von {todos.length} erledigt
            </span>
            <progress
              value={completed}
              max={todos.length || 1}
              aria-label="Erledigte Aufgaben"
            />
          </footer>
        </section>
        <p className={styles.note}>
          Deine Aufgaben bleiben bis zum nächsten Serverneustart erhalten.
        </p>
      </div>
    </main>
  );
}
