import { useState, useEffect } from 'react'

import { Button } from '@/components/ui/button'
import { Pagination, PaginationContent, PaginationEllipsis, PaginationItem, PaginationLink, PaginationNext, PaginationPrevious } from '@/components/ui/pagination'
import { ChevronLeft, ChevronRight } from 'lucide-react'
import { cn } from '@/lib/utils'
import { useRouter } from 'next/navigation'

interface Props {
  page: number
  pageSize: number
  pathname?: string
  isLink?: boolean
  onClick?: (pageNumber: number) => void
}

const RANGE = 1

export default function AutoPagination({
  page,
  pageSize,
  pathname = '/',
  isLink = true,
  onClick = () => {}
}: Props) {
  const [mounted, setMounted] = useState(false)
  const router = useRouter()

  useEffect(() => {
    setMounted(true)
  }, [])

  const handlePageChange = (pageNumber: number) => {
//     if (isLink && mounted) {
//   router.push({
//   pathname,
//   query: { ...router.query, page: pageNumber.toString() }
// })
//     } else {
//       onClick(pageNumber)
//     }
  }

  const renderPagination = () => {
    let dotAfter = false
    let dotBefore = false
    const renderDotBefore = (index: number) => {
      if (!dotBefore) {
        dotBefore = true
        return (
          <PaginationItem key={`before-${index}`}>
            <PaginationEllipsis />
          </PaginationItem>
        )
      }
      return null
    }
    const renderDotAfter = (index: number) => {
      if (!dotAfter) {
        dotAfter = true
        return (
          <PaginationItem key={`after-${index}`}>
            <PaginationEllipsis />
          </PaginationItem>
        )
      }
      return null
    }
    return Array(pageSize)
      .fill(0)
      .map((_, index) => {
        const pageNumber = index + 1

        if (
          page <= RANGE * 2 + 1 &&
          pageNumber > page + RANGE &&
          pageNumber < pageSize - RANGE + 1
        ) {
          return renderDotAfter(index)
        } else if (page > RANGE * 2 + 1 && page < pageSize - RANGE * 2) {
          if (pageNumber < page - RANGE && pageNumber > RANGE) {
            return renderDotBefore(index)
          } else if (
            pageNumber > page + RANGE &&
            pageNumber < pageSize - RANGE + 1
          ) {
            return renderDotAfter(index)
          }
        } else if (
          page >= pageSize - RANGE * 2 &&
          pageNumber > RANGE &&
          pageNumber < page - RANGE
        ) {
          return renderDotBefore(index)
        }
        return (
          <PaginationItem key={index}>
            {isLink ? (
              <PaginationLink
                href="#"
                onClick={(e) => {
                  e.preventDefault()
                  handlePageChange(pageNumber)
                }}
                isActive={pageNumber === page}
              >
                {pageNumber}
              </PaginationLink>
            ) : (
              <Button
                onClick={() => handlePageChange(pageNumber)}
                variant={pageNumber === page ? 'outline' : 'ghost'}
                className="w-9 h-9 p-0"
              >
                {pageNumber}
              </Button>
            )}
          </PaginationItem>
        )
      })
  }

  if (!mounted) return null

  return (
    <Pagination>
      <PaginationContent>
        <PaginationItem>
          {isLink ? (
            <PaginationPrevious
              href="#"
              className={cn({
                'cursor-not-allowed': page === 1
              })}
              onClick={(e) => {
                e.preventDefault()
                if (page !== 1) handlePageChange(page - 1)
              }}
            />
          ) : (
            <Button
              disabled={page === 1}
              className="h-9 p-0 px-2"
              variant="ghost"
              onClick={() => handlePageChange(page - 1)}
            >
              <ChevronLeft className="w-5 h-5" /> Previous
            </Button>
          )}
        </PaginationItem>
        {renderPagination()}
        <PaginationItem>
          {isLink ? (
            <PaginationNext
              href="#"
              className={cn({
                'cursor-not-allowed': page === pageSize
              })}
              onClick={(e) => {
                e.preventDefault()
                if (page !== pageSize) handlePageChange(page + 1)
              }}
            />
          ) : (
            <Button
              disabled={page === pageSize}
              className="h-9 p-0 px-2"
              variant="ghost"
              onClick={() => handlePageChange(page + 1)}
            >
              Next <ChevronRight className="w-5 h-5" />
            </Button>
          )}
        </PaginationItem>
      </PaginationContent>
    </Pagination>
  )
}