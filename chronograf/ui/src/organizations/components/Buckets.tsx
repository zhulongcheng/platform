// Libraries
import React, {PureComponent, ChangeEvent} from 'react'
import _ from 'lodash'

// Components
import IndexList from 'src/shared/components/index_views/IndexList'
import {
  Input,
  Alignment,
  ComponentSize,
  EmptyState,
  Button,
  ComponentColor,
  IconFont,
} from 'src/clockface'
import ProfilePage from 'src/shared/components/profile_page/ProfilePage'

// Types
import {Bucket} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

interface Props {
  buckets: Bucket[]
}

interface State {
  filterTerm: string
}

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

@ErrorHandling
export default class Buckets extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      filterTerm: '',
    }
  }

  public render() {
    const {filterTerm} = this.state
    return (
      <>
        <ProfilePage.Header>
          <Input
            icon={IconFont.Search}
            placeholder="Filter Buckets..."
            widthPixels={290}
            value={filterTerm}
            onChange={this.handleFilterChange}
            onBlur={this.handleFilterBlur}
          />
          <Button
            text="Create Bucket"
            icon={IconFont.Plus}
            color={ComponentColor.Primary}
          />
        </ProfilePage.Header>
        <IndexList
          columns={this.columns}
          rows={this.rows}
          emptyState={this.emptyState}
        />
      </>
    )
  }

  private get columns(): IndexListColumn[] {
    return [
      {
        key: 'bucket--name',
        title: 'Name',
        size: 500,
        showOnHover: false,
        align: Alignment.Left,
      },
      {
        key: 'bucket--retention',
        title: 'Retention Period',
        size: 100,
        showOnHover: false,
        align: Alignment.Left,
      },
    ]
  }

  private get filteredBuckets(): Bucket[] {
    const {buckets} = this.props
    const {filterTerm} = this.state

    const matchingBuckets = buckets.filter(b =>
      b.name.toLowerCase().includes(filterTerm.toLowerCase())
    )

    return _.sortBy(matchingBuckets, b => b.name.toLowerCase())
  }

  private get rows(): IndexListRow[] {
    return this.filteredBuckets.map(bucket => ({
      disabled: false,
      columns: [
        {
          key: 'bucket--name',
          contents: <a href="#">{bucket.name}</a>,
        },
        {
          key: 'bucket--retention',
          contents: bucket.retentionPeriod,
        },
      ],
    }))
  }

  private get emptyState(): JSX.Element {
    const {filterTerm} = this.state

    if (_.isEmpty(filterTerm)) {
      return (
        <EmptyState size={ComponentSize.Large}>
          <EmptyState.Text text="Oh noes I dun see na buckets" />
        </EmptyState>
      )
    }

    return (
      <EmptyState size={ComponentSize.Large}>
        <EmptyState.Text text="No buckets match your query" />
      </EmptyState>
    )
  }

  private handleFilterBlur = (e: ChangeEvent<HTMLInputElement>): void => {
    this.setState({filterTerm: e.target.value})
  }

  private handleFilterChange = (e: ChangeEvent<HTMLInputElement>): void => {
    this.setState({filterTerm: e.target.value})
  }
}
