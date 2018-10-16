// Libraries
import React, {PureComponent} from 'react'
import {Link} from 'react-router'

// Components
import IndexList from 'src/shared/components/index_views/IndexList'
import {Alignment, ComponentSize, EmptyState} from 'src/clockface'

// Types
import {Dashboard} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

interface Props {
  dashboards: Dashboard[]
}

@ErrorHandling
export default class Dashboards extends PureComponent<Props> {
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
        key: 'dashboard--name',
        title: 'Name',
        size: 500,
        showOnHover: false,
        align: Alignment.Left,
      },
      {
        key: 'dashboard--actions',
        title: '',
        size: 100,
        showOnHover: true,
        align: Alignment.Right,
      },
    ]
  }

  private get rows(): IndexListRow[] {
    const {dashboards} = this.props

    return dashboards.map(dashboard => ({
      disabled: false,
      columns: [
        {
          key: 'dashboard--name',
          contents: (
            <Link to={`/dashboards/${dashboard.id}`}>{dashboard.name}</Link>
          ),
        },
        {
          key: 'dashboard--actions',
          contents: <p>DELETE</p>,
        },
      ],
    }))
  }

  private get emptyState(): JSX.Element {
    return (
      <EmptyState size={ComponentSize.Large}>
        <EmptyState.Text text="Oh noes I dun see na dashbardsss" />
      </EmptyState>
    )
  }
}
