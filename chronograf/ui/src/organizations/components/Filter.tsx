import {PureComponent} from 'react'
import _ from 'lodash'

interface Props<T> {
  list: T[]
  searchTerm: string
  searchKeys: string[]
  children: (list: T[]) => any
}

export default class FilterList<T> extends PureComponent<Props<T>> {
  public render() {
    return this.props.children(this.filtered)
  }

  private get filtered(): T[] {
    const {list, searchKeys, searchTerm} = this.props

    const filtered = list.filter(item => {
      const isInList = Object.entries(item).some(([key, value]) => {
        if (searchKeys.includes(key)) {
          return String(value)
            .toLocaleLowerCase()
            .includes(searchTerm.toLocaleLowerCase())
        }
      })

      return isInList
    })

    return filtered
  }
}
