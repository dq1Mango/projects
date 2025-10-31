mod logger;
mod multi_line_string;
#[cfg(test)]
mod tests;

mod update;

use std::{collections::HashMap, fmt::Debug, hash::Hash, rc::Rc, time::Duration, vec};

use chrono::{DateTime, TimeDelta, Utc};
use ratatui::{
  Frame,
  buffer::Buffer,
  layout::{Constraint, Direction, Layout, Position, Rect},
  style::{Color, Modifier, Style, Stylize},
  symbols::border,
  text::{Line, Span},
  widgets::{Block, Paragraph, Widget},
};
// use ratatui_image::{StatefulImage, picker::Picker, protocol::StatefulProtocol};

use crate::logger::Logger;
use crate::multi_line_string::MultiLineString;
use crate::update::*;

// #[derive(Debug, Default)]
pub struct Model {
  running_state: RunningState,
  mode: Mode,
  contacts: Contacts,
  // groups: Vec<Group,
  chats: Vec<Chat>,
  chat_index: usize,
}

#[derive(Debug, Default, PartialEq, Eq)]
pub enum RunningState {
  #[default]
  Running,
  OhShit,
}

#[derive(Default, Debug, PartialEq)]
pub enum Mode {
  #[default]
  Normal,
  Insert,
}

// #[derive(Debug, Default)]
// pub struct TimeStamps {
//   sent: DateTime<Utc>,
//   recieved: Option<DateTime<Utc>>,
//   readby: Option<Vec<(Contact, DateTime<Utc>)>>,
// }

#[derive(Debug)]
pub struct NotMyMessage {
  sender: PhoneNumber,
  sent: DateTime<Utc>,
}

#[derive(Debug)]
pub struct MyMessage {
  sent: DateTime<Utc>,
  // these r kind of a mess
  delivered_to: Vec<(PhoneNumber, Option<DateTime<Utc>>)>,
  read_by: Vec<(PhoneNumber, Option<DateTime<Utc>>)>,
}

#[derive(Debug)]
pub enum Metadata {
  MyMessage(MyMessage),
  NotMyMessage(NotMyMessage),
}

#[derive(Debug)]
pub struct Message {
  body: MultiLineString,
  metadata: Metadata,
}

#[derive(Default, Debug)]
pub struct Location {
  index: usize,
  offset: i16,
  requested_scroll: i16,
}

// pub struct MyImageWrapper(StatefulProtocol);

// sshhhhhh
// impl Debug for MyImageWrapper {
//   fn fmt(&self, _f: &mut fmt::Formatter<'_>) -> fmt::Result {
//     Ok(())
//   }
// }

#[derive(Hash, PartialEq, Eq, Debug)]
struct PhoneNumber(String);

impl Clone for PhoneNumber {
  fn clone(&self) -> Self {
    PhoneNumber(self.0.clone())
  }
}

#[derive(Debug, Default)]
struct Group {
  name: String,
  // icon: Option<MyImageWrapper>,
  members: Vec<PhoneNumber>,
  _description: String,
}

#[derive(Debug, Default)]
pub struct Contact {
  _name: String,
  nick_name: String,
  // pfp: Option<MyImageWrapper>,
  // icon: Image,
}

type Contacts = Rc<HashMap<PhoneNumber, Contact>>;

#[derive(Debug, Default)]
pub struct Chat {
  // id: u8,
  participants: Group,
  messages: Vec<Message>,
  location: Location,
  text_input: TextInput,
}

#[derive(Debug, Default)]
pub struct TextInput {
  body: MultiLineString,
  cursor_index: u16,
  cursor_position: Position,
}

pub struct Settings {
  borders: bool,
  message_width_ratio: f32,
  _identity: String,
}

impl Settings {
  fn init() -> Self {
    Self {
      borders: true,
      message_width_ratio: 0.8,
      _identity: "me".to_string(),
    }
  }
}

impl Model {
  fn init() -> Self {
    let dummy_number = PhoneNumber("14124206767".to_string());

    let messages = vec![
      Message {
        body: MultiLineString::init(
          "first message lets make this   message super looong jjafkldjaflk it was not long enough last time time to yap fr",
        ),
        metadata: Metadata::NotMyMessage(NotMyMessage {
          sender: dummy_number.clone(),
          sent: Utc::now().checked_sub_signed(TimeDelta::minutes(2)).expect("kaboom"),
        }),
      },
      Message {
        body: MultiLineString::init("second message"),
        metadata: Metadata::MyMessage(MyMessage {
          sent: Utc::now(),
          read_by: vec![(dummy_number.clone(), Some(Utc::now()))],
          delivered_to: vec![(dummy_number.clone(), None)],
        }),
      },
      Message {
        body: MultiLineString::init("a luxurious third message because im not convinced yet"),
        metadata: Metadata::MyMessage(MyMessage {
          sent: Utc::now(),
          read_by: vec![(dummy_number.clone(), None)],
          delivered_to: vec![(dummy_number.clone(), None)],
        }),
      },
    ];

    let mut chat = Chat::default();

    for message in messages {
      chat.messages.push(message);
    }

    // let picker = Picker::from_query_stdio().expect("kaboom");

    // Load an image with the image crate.
    // let dyn_img = image::ImageReader::open("./assets/ferris_the_wheel.jpg")
    //   .unwrap()
    //   .decode()
    //   .unwrap();

    // Create the Protocol which will be used by the widget.
    // let image = picker.new_resize_protocol(dyn_img.clone());
    // let image2 = picker.new_resize_protocol(dyn_img);

    chat.participants = Group {
      members: vec![dummy_number.clone()],
      name: "group 1".to_string(),
      // icon: Some(MyImageWrapper(image)),
      _description: "".to_string(),
    };
    chat.text_input = TextInput::default();
    chat.location = Location {
      index: 1,
      offset: 0,
      requested_scroll: 0,
    };
    // let chats: Vec<Chat> = vec![chat];

    let mut contacts = HashMap::new();

    contacts.insert(
      dummy_number,
      Contact {
        nick_name: String::from("nickname"),
        _name: String::from("name"),
        // pfp: Some(MyImageWrapper(image2)),
      },
    );

    let model = Model {
      chat_index: 0,
      contacts: Rc::new(contacts),
      chats: vec![chat],
      running_state: RunningState::Running,
      mode: Mode::Normal,
    };
    // let mut model = Model::default();
    // model.contacts = contacts;
    // // model.chats = chats;
    // model.chat_index = 0;
    model
  }

  // not really needed but it staves off the need for explicit liiftimes a little longer
  fn current_chat(&mut self) -> &mut Chat {
    &mut self.chats[self.chat_index]
  }
}

impl TextInput {
  fn render(&mut self, area: Rect, buf: &mut Buffer) {
    let block = Block::bordered().border_set(border::THICK);

    // shitty temp padding for the border
    // let mut area = area;
    // area.x += 1;
    // area.width -= 2;
    // area.height -= 2;
    // area.y += 1;

    // minus 3 b/c you cant have the cursor on the border and i cant be bothered to add another
    // edge case
    let vec_lines = self.body.as_trimmed_lines(area.width - 3);
    // logger.log(format!("this is the first line: {}", self.cursor_index));
    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(area, buf);

    self.cursor_position = self.calc_cursor_position(area)
  }

  fn calc_cursor_position(&mut self, area: Rect) -> Position {
    // gotta pad the border (still havent found a better way of doing this)
    let mut pos = Position {
      x: area.x + 1,
      y: area.y + 1,
    };
    // mad ugly calculations, smthns gotta change
    let lines = self.body.as_lines(area.width - 3);
    // let body = self.body.body.char_indices();

    let (mut index, mut row) = (0, 0);

    while (index + lines[row].len() as u16) < self.cursor_index {
      index += lines[row].len() as u16;
      pos.y += 1;
      row += 1;
    }

    pos.x += (self.cursor_index - index).clamp(0, area.width - 3);
    pos
  }

  fn insert_char(&mut self, char: char, _logger: &mut Logger) {
    // some disgusting object-oriented blashphemy going on here
    self.body.body.insert(self.cursor_index as usize, char);
    self.cursor_index += 1;
  }

  fn delete_char(&mut self) {
    if self.cursor_index == 0 {
      return;
    }

    self.cursor_index -= 1;
    self.body.body.remove(self.cursor_index as usize);
  }
}

impl Message {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, contacts: &Contacts) {
    let mut block = Block::bordered().border_set(border::THICK);

    if let Metadata::NotMyMessage(x) = &self.metadata {
      block = block.title_top(Line::from(contacts[&x.sender].nick_name.clone()).left_aligned());
    }
    // this ugly shadow cost me a good 15 mins of my life ... but im not changing it
    let mut my_area = area.clone();
    my_area.width = (area.width as f32 * settings.message_width_ratio + 0.5) as u16;
    // let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;

    let vec_lines: Vec<String> = self.body.as_trimmed_lines(my_area.width - 2);

    // shrink the message to fit if it does not need mutliple lines
    if vec_lines.len() == 1 {
      my_area.width = vec_lines[0].len() as u16 + 2;
    }

    // "allign" the chat to the right if it was sent by you
    // TODO: should add setting to toggle this behavior

    // look at this cool syntax i learned today
    if let Metadata::MyMessage(_) = self.metadata {
      my_area.x += area.width - my_area.width;
    }

    let mut lines: Vec<Line> = Vec::new();
    for yap in vec_lines {
      lines.push(Line::from(yap));
    }

    Paragraph::new(lines).block(block).render(my_area, buf)
    // .wrap(Wrap { trim: true })
  }

  // i thought i knew how lifetimes worked
  fn format_delivered_status(&self) -> Line<'_> {
    let check_icon = "";

    return match &self.metadata {
      Metadata::NotMyMessage(_) => Line::from(""),
      Metadata::MyMessage(x) => {
        if x.all_read() {
          Line::from(Span::styled(
            [check_icon, " ", check_icon].concat(),
            Style::default().fg(Color::White).add_modifier(Modifier::BOLD),
          ))
        } else if x.all_delivered() {
          Line::from(Span::styled(
            [check_icon, " ", check_icon].concat(),
            Style::default().fg(Color::Gray),
          ))
        } else if x.sent() {
          Line::from(Span::styled(check_icon, Style::default().fg(Color::Gray)))
        } else {
          Line::from(Span::styled("_", Style::default().fg(Color::White)))
        }
      }
    };
  }

  fn height(&mut self, width: u16) -> u16 {
    self.body.as_lines(width).len() as u16 + 2
  }
}

fn _format_vec(vec: &Vec<String>) -> String {
  let mut output = String::from("[");

  for thing in vec {
    output.push_str(thing);
    output.push_str(", ");
  }

  output.push_str("]");

  return output;
}

impl Chat {
  fn render(&mut self, area: Rect, buf: &mut Buffer, settings: &Settings, contacts: Contacts) {
    let input_lines = self.text_input.body.rows(area.width - 3);
    // Logger::log("this is our input: ".to_string());
    // Logger::log(format_vec(self.text_input.body.as_lines(area.width - 2)));

    let layout = Layout::vertical([Constraint::Min(6), Constraint::Length(input_lines + 2)]).split(area);

    // kind of a sketchy shadow here but the layout[1] is used like once
    let area = layout[0];

    let block = Block::bordered().border_set(border::THICK);
    // .title(title.centered())
    // .title_bottom(instructions.centered())
    block.render(area, buf);

    // shitty temp padding for the border
    let mut area = area;
    area.x += 1;
    area.width -= 2;
    area.height -= 2;
    area.y += 1;
    // end shitty tmp padding

    let message_width: u16 = (area.width as f32 * settings.message_width_ratio + 0.5) as u16 - 2;

    let mut scroll = self.location.requested_scroll;
    let mut index = self.location.index;
    let mut offset = self.location.offset;

    // yeah this scrolling logic is a little ugly but im not sure how to make it less so
    // also im a little scared to touch it
    if scroll > 0 {
      while scroll > 0 {
        if index + 1 == self.messages.len() {
          offset = 0;
          break;
        }

        let height = self.messages[index + 1].height(message_width);

        if height as i16 > scroll + offset {
          offset += scroll;
          break;
        }
        index += 1;
        scroll -= height as i16;

        if scroll < 0 {
          offset += scroll;
          scroll = 0;
        }
      }
    } else if scroll < 0 {
      while scroll < 0 {
        if offset as i16 >= scroll * -1 {
          offset += scroll;
          break;
        }
        if index == 0 {
          offset = 0;
          break;
        }

        let height = self.messages[index].height(message_width);
        scroll += height as i16;
        index -= 1;

        if scroll > 0 {
          offset = scroll;
          scroll = 0;
        }
      }
    }

    self.location.index = index;
    self.location.offset = offset;
    self.location.requested_scroll = 0;

    let mut y = area.height as i16 - self.location.offset;

    loop {
      let message = &mut self.messages[index];

      let height = message.body.rows(message_width) + 2;

      y -= height as i16;
      if y < 0 {
        break;
      }

      // let height = min(y + requested_height, area.height);
      let new_area = Rect::new(area.x, area.y + y as u16, area.width, height as u16);

      message.render(new_area, buf, settings, &contacts);

      if index == 0 {
        break;
      }

      index -= 1;
    }

    self.text_input.render(layout[1], buf);
  }

  fn last_message(&self) -> &Message {
    let last = self.messages.len() - 1;
    &self.messages[last]
  }
}

trait MyStringUtils {
  fn shrink<T>(&self, width: T) -> String
  where
    T: Into<usize>;
}

impl MyStringUtils for String {
  fn shrink<T>(&self, width: T) -> String
  where
    T: Into<usize>,
  {
    let width = width.into();

    let mut fitted = self.clone();

    if fitted.len() <= width {
      return fitted;
    } else {
      fitted = fitted[..width - 3].to_string();
      fitted.push_str("...");
      // fitted.push("...");
      return fitted;
    }
  }
}

fn format_duration(message: &Message) -> String {
  let time: DateTime<Utc>;

  match &message.metadata {
    Metadata::NotMyMessage(x) => time = x.sent,
    Metadata::MyMessage(x) => time = x.sent,
  }

  let duration = Utc::now().signed_duration_since(time);

  if duration.num_minutes() < 1 {
    return "Now".to_string();
  } else if duration.num_hours() < 1 {
    let mut temp = duration.num_minutes().to_string();
    temp.push_str("m");
    return temp;
  } else if duration.num_days() < 1 {
    let mut temp = duration.num_hours().to_string();
    temp.push_str("h");
    return temp;
  } else {
    return time.format("%M %D").to_string();
  }

  // let mut result = num.to_string();
  // result.push_str(chr);
  // result
}

impl MyMessage {
  fn all_read(&self) -> bool {
    for (_, date) in &self.read_by {
      match date {
        Some(_) => {}
        None => return false,
      }
    }

    true
  }

  fn all_delivered(&self) -> bool {
    for (_, date) in &self.delivered_to {
      match date {
        Some(_) => {}
        None => return false,
      }
    }

    true
  }

  fn sent(&self) -> bool {
    true
  }
}

fn render_group(chat: &mut Chat, area: Rect, buf: &mut Buffer) {
  // let icon = &mut chat.participants.icon;

  Block::bordered().border_set(border::THICK).render(area, buf);

  let mut area = area;
  area.x += 1;
  area.width -= 2;
  area.height -= 2;
  area.y += 1;

  let layout = Layout::horizontal([Constraint::Length(7), Constraint::Min(15), Constraint::Length(6)]).split(area);

  // let image = StatefulImage::default().resize(Resize::Crop(None));
  // let mut pfp = match &self.pfp {
  //   Some(x) => x.0,
  //   None => panic!("Aaaaaahhhhh"),
  // };
  // // StatefulImage::render(image, layout[0], buf, &mut pfp);
  // let image: StatefulImage<StatefulProtocol> = StatefulImage::default();

  // match icon.as_mut() {
  // Some(image) => StatefulImage::new().render(area, buf, &mut image.0),
  // None => {}
  // }

  let last_message = chat.last_message();
  let group = &chat.participants;

  let message_text: Vec<String> = last_message.body.fit(layout[1].width, layout[1].height - 1);

  let mut innner_lines: Vec<Line> = vec![Line::from(group.name.shrink(layout[1].width).bold())];

  for line in message_text {
    innner_lines.push(Line::from(line));
  }

  Paragraph::new(innner_lines).render(layout[1], buf);

  let time = format_duration(last_message);

  Paragraph::new(vec![Line::from(time), last_message.format_delivered_status()]).render(layout[2], buf);
}

// impl Group {
//   fn render(&mut self, last_message: &Message, area: Rect, buf: &mut Buffer) {
//     Block::bordered().border_set(border::THICK).render(area, buf);
//
//     let mut area = area;
//     area.x += 1;
//     area.width -= 2;
//     area.height -= 2;
//     area.y += 1;
//
//     let layout = Layout::horizontal([Constraint::Length(7), Constraint::Min(15), Constraint::Length(6)]).split(area);
//
//     // let image = StatefulImage::default().resize(Resize::Crop(None));
//     // let mut pfp = match &self.pfp {
//     //   Some(x) => x.0,
//     //   None => panic!("Aaaaaahhhhh"),
//     // };
//     // // StatefulImage::render(image, layout[0], buf, &mut pfp);
//     // let image: StatefulImage<StatefulProtocol> = StatefulImage::default();
//     StatefulImage::new().render(area, buf, &mut self.icon.as_mut().unwrap().0);
//     let message_text: Vec<String> = last_message.body.fit(layout[1].width, layout[1].height - 1);
//
//     let mut innner_lines: Vec<Line> = vec![Line::from(self.name.shrink(layout[1].width).bold())];
//
//     for line in message_text {
//       innner_lines.push(Line::from(line));
//     }
//
//     Paragraph::new(innner_lines).render(layout[1], buf);
//
//     let time = format_duration(last_message);
//
//     Paragraph::new(vec![Line::from(time), last_message.format_delivered_status()]).render(layout[2], buf);
//   }
// }
// /

// main ---
fn main() -> color_eyre::Result<()> {
  // tui::install_panic_hook();
  let mut terminal = ratatui::init();
  let mut model = Model::init();
  let settings = &Settings::init();

  // regular lumber jack
  let logger = &mut Logger::init("log.txt");
  Logger::log("testing".to_string());

  while model.running_state != RunningState::OhShit {
    // Render the current view
    terminal.draw(|f| view(&mut model, f, settings))?;

    // Handle events and map to a Message
    let mut current_msg = handle_event(&model)?;

    // Process updates as long as they return a non-None message
    while current_msg.is_some() {
      current_msg = update(&mut model, current_msg.unwrap(), logger);
    }
  }

  ratatui::restore();
  Ok(())
}

// TODO: gotta figure out how to model chat state

fn view(model: &mut Model, frame: &mut Frame, settings: &Settings) {
  let title = Line::from(" Counter App Tutorial ".bold());
  let instructions = Line::from(vec![
    " Decrement ".into(),
    "<Left>".blue().bold(),
    " Increment ".into(),
    "<Right>".blue().bold(),
    " Quit ".into(),
    "<Q> ".blue().bold(),
  ]);
  let _block = Block::bordered()
    .title(title.centered())
    .title_bottom(instructions.centered())
    .border_set(border::THICK);

  // let _counter_text = Text::from(vec![Line::from(vec![
  //   "Value: ".into(),
  //   model.counter.to_string().yellow(),
  // ])]);

  let layout = Layout::new(
    Direction::Horizontal,
    vec![Constraint::Percentage(40), Constraint::Percentage(60)],
  )
  .split(frame.area());

  _ = Block::bordered()
    .border_set(border::THICK)
    .render(layout[0], frame.buffer_mut());

  let contact_height = 3 + 2;

  let mut contact_area = layout[0];
  contact_area.width -= 2;
  contact_area.height = contact_height;
  contact_area.x += 1;
  contact_area.y += 1;

  let mut index = 0;

  while contact_area.y < layout[0].height && index < model.chats.len() {
    let chat = &mut model.chats[index];
    render_group(chat, contact_area, frame.buffer_mut());
    // let last = &(&mut model.chats)[index].last_message();
    // model.chats[index].participants.render(last, contact_area, frame.buffer_mut());
    contact_area.y += contact_height;
    index += 1;
  }

  let contacts = Rc::clone(&model.contacts);
  model
    .current_chat()
    .render(layout[1], frame.buffer_mut(), settings, contacts);

  frame.set_cursor_position(model.current_chat().text_input.cursor_position);

  //
  // frame.render_widget(
  //   Paragraph::new(message_text).right_aligned().block(block),
  //   frame.area(),
  // );

  // let p = Paragraph::new("A very long text that might not fit the container...")
  //   .wrap(Wrap { trim: true });
  //
  // let test_rect = Rect::new(10, 10, 7, 20);
  // frame.render_widget(p, test_rect);
}
