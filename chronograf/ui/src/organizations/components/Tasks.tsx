// Libraries
import React, {PureComponent} from 'react'
import {Link} from 'react-router'

// Components
import IndexList from 'src/shared/components/index_views/IndexList'
import {Alignment, ComponentSize, EmptyState} from 'src/clockface'

// Types
import {Task} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

interface Props {
  tasks: Task[]
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
        key: 'task--name',
        title: 'Name',
        size: 500,
        showOnHover: false,
        align: Alignment.Left,
      },
      {
        key: 'task--actions',
        title: '',
        size: 100,
        showOnHover: true,
        align: Alignment.Right,
      },
    ]
  }

  private get rows(): IndexListRow[] {
    const {tasks} = this.props

    if (!tasks) {
      return []
    }

    return tasks.map(task => ({
      disabled: false,
      columns: [
        {
          key: 'task--name',
          contents: <Link to={`/tasks/${task.id}`}>{task.name}</Link>,
        },
        {
          key: 'task--actions',
          contents: 'DELETE',
        },
      ],
    }))
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Medium}>
        <EmptyState.Text text="I see nay a task" />
      </EmptyState>
    )
  }
}
