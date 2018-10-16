// Libraries
import React, {PureComponent} from 'react'

// Components
import IndexList from 'src/shared/components/index_views/IndexList'
import {Alignment, ComponentSize, EmptyState} from 'src/clockface'

// Types
import {Member} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

interface Props {
  members: Member[]
}

export default class Members extends PureComponent<Props> {
  public render() {
    return (
      <IndexList
        columns={this.columns}
        rows={this.rows}
        emptyState={this.emptyState}
      />
    )
  }

  private get columns(): IndexListColumn[] {
    return [
      {
        key: 'member--name',
        title: 'Name',
        size: 500,
        showOnHover: false,
        align: Alignment.Left,
      },
      {
        key: 'member--actions',
        title: '',
        size: 100,
        showOnHover: true,
        align: Alignment.Right,
      },
    ]
  }

  private get rows(): IndexListRow[] {
    const {members} = this.props

    if (!members) {
      return []
    }

    return members.map(member => ({
      disabled: false,
      columns: [
        {
          key: 'member--name',
          contents: <a href="#">{member.name}</a>,
        },
        {
          key: 'member--actions',
          contents: 'REMOVE',
        },
      ],
    }))
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Medium}>
        <EmptyState.Text text="This org has been abandoned" />
      </EmptyState>
    )
  }
}
