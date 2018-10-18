// Libraries
import React, {PureComponent, ChangeEvent} from 'react'

// Components
import {Form, Input, InputType, Columns} from 'src/clockface'

interface Props {
  retentionPeriod: number
  onChangeRetentionPeriod: (rp: number) => void
}

interface Time {
  days: number
  hours: number
  minutes: number
  seconds: number
}

enum TimeKey {
  Days = 'days',
  Hours = 'hours',
  Minutes = 'minutes',
  Seconds = 'seconds',
}

interface State {
  time: Time
}

export default class RetentionPeriod extends PureComponent<Props, State> {
  public render() {
    const {days, hours, minutes, seconds} = this.msToTime

    return (
      <>
        <Form.Element label="Days" colsXS={Columns.Three}>
          <Input
            name={TimeKey.Days}
            type={InputType.Number}
            value={`${days}`}
            onChange={this.handleChangeInput}
          />
        </Form.Element>
        <Form.Element label="Hours" colsXS={Columns.Three}>
          <Input
            name={TimeKey.Hours}
            min="0"
            type={InputType.Number}
            value={`${hours}`}
            onChange={this.handleChangeInput}
          />
        </Form.Element>
        <Form.Element label="Minutes" colsXS={Columns.Three}>
          <Input
            name={TimeKey.Minutes}
            min="0"
            type={InputType.Number}
            value={`${minutes}`}
            onChange={this.handleChangeInput}
          />
        </Form.Element>
        <Form.Element label="Seconds" colsXS={Columns.Three}>
          <Input
            name={TimeKey.Seconds}
            min="0"
            type={InputType.Number}
            value={`${seconds}`}
            onChange={this.handleChangeInput}
          />
        </Form.Element>
      </>
    )
  }

  private timeToMs = ({days, hours, minutes, seconds}: Time): number => {
    const msInSecond = 1000
    const msInMinute = msInSecond * 60
    const msInHour = msInMinute * 60
    const msInDay = msInHour * 24

    const msDays = msInDay * days
    const msHours = msInHour * hours
    const msMinutes = msInMinute * minutes
    const msSeconds = msInSecond * seconds

    const retentionPeriod = msDays + msHours + msMinutes + msSeconds
    return retentionPeriod
  }

  private get msToTime(): Time {
    const {retentionPeriod} = this.props
    let seconds = Math.floor(retentionPeriod / 1000)
    let minutes = Math.floor(seconds / 60)
    seconds = seconds % 60
    let hours = Math.floor(minutes / 60)
    minutes = minutes % 60
    const days = Math.floor(hours / 24)
    hours = hours % 24

    return {
      days,
      hours,
      minutes,
      seconds,
    }
  }

  private handleChangeInput = (e: ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value
    const key = e.target.name as keyof Time
    const time = {...this.msToTime, [key]: Number(value)}
    const ms = this.timeToMs(time)

    this.props.onChangeRetentionPeriod(ms)
  }
}
