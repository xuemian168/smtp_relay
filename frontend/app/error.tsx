'use client'

export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string }
  reset: () => void
}) {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-4xl font-bold text-red-600">
          出错了 / Something went wrong
        </h1>
        <p className="mt-2 text-lg text-gray-600 dark:text-gray-400">
          {error.message || '发生了未知错误 / An unknown error occurred'}
        </p>
        <button
          onClick={() => reset()}
          className="mt-4 inline-block rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-500"
        >
          重试 / Try again
        </button>
      </div>
    </div>
  )
} 