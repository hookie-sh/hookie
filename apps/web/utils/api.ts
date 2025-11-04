export async function fetcher<T>(url: string): Promise<T> {
  const res = await fetch(url);
  
  if (!res.ok) {
    const error = await res.json().catch(() => ({ error: "An error occurred" }));
    throw new Error(error.error || `Failed to fetch: ${res.statusText}`);
  }
  
  return res.json();
}






