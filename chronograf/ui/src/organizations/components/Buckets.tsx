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
  OverlayTechnology,
} from 'src/clockface'
import ProfilePage from 'src/shared/components/profile_page/ProfilePage'
import BucketOverlay from 'src/organizations/components/BucketOverlay'

// APIs
import {createBucket} from 'src/organizations/apis'

// Types
import {Bucket, Organization} from 'src/types/v2'
import {
  IndexListColumn,
  IndexListRow,
} from 'src/shared/components/index_views/IndexListTypes'

interface Props {
  org: Organization
  buckets: Bucket[]
}

interface State {
  buckets: Bucket[]
  filterTerm: string
  modalState: ModalState
}

enum ModalState {
  Open = 'open',
  Closed = 'closed',
}

// Decorators
import {ErrorHandling} from 'src/shared/decorators/errors'

@ErrorHandling
export default class Buckets extends PureComponent<Props, State> {
  constructor(props: Props) {
    super(props)

    this.state = {
      filterTerm: '',
      buckets: this.props.buckets,
      modalState: ModalState.Closed,
    }
  }

  public render() {
    const {org} = this.props
    const {filterTerm, modalState} = this.state

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
            onClick={this.handleOpenModal}
          />
        </ProfilePage.Header>
        <IndexList
          rows={this.rows}
          columns={this.columns}
          emptyState={this.emptyState}
        />
        <OverlayTechnology visible={modalState === ModalState.Open}>
          <BucketOverlay
            link={org.links.buckets}
            onCloseModal={this.handleCloseModal}
            onCreateBucket={this.handleCreateBucket}
          />
        </OverlayTechnology>
      </>
    )
  }

  private handleCreateBucket = async (
    link: string,
    bucket: Partial<Bucket>
  ): Promise<void> => {
    const {buckets} = this.state
    const b = await createBucket(link, bucket)
    this.setState({buckets: [b, ...buckets]})
    this.handleCloseModal()
  }

  private handleOpenModal = (): void => {
    this.setState({modalState: ModalState.Open})
  }

  private handleCloseModal = (): void => {
    this.setState({modalState: ModalState.Closed})
  }

  private handleFilterBlur = (e: ChangeEvent<HTMLInputElement>): void => {
    this.setState({filterTerm: e.target.value})
  }

  private handleFilterChange = (e: ChangeEvent<HTMLInputElement>): void => {
    this.setState({filterTerm: e.target.value})
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
}
