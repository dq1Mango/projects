use std::time::Duration;

use ratatui::{
  Frame,
  buffer::Buffer,
  crossterm::event::{self, Event, KeyCode},
  layout::Rect,
  style::Stylize,
  symbols::border,
  text::{Line, Text},
  widgets::{Block, Paragraph, Widget, Wrap},
};

#[derive(Debug, Default)]
struct Model {
  counter: i32,
  running_state: RunningState,
  content: Chat,
}

#[derive(Debug, Default, PartialEq, Eq)]
enum RunningState {
  #[default]
  Running,
  Done,
}

#[derive(PartialEq)]
enum Action {
  Increment,
  Decrement,
  Reset,
  Quit,
}

#[derive(Debug, Default, Clone)]
pub struct Message {
  content: String,
  sender: String,
}

#[derive(Debug, Default)]
pub struct Chat {
  id: u8,
  data: Vec<Message>,
}

impl Model {
  fn init() -> Model {
    let messages = vec![
      Message {
        content: String::from("first message"),
        sender: String::from(""),
      },
      Message {
        content: String::from("second message"),
        sender: String::from(""),
      },
    ];

    let mut chat = Chat::default();

    for message in messages {
      chat.data.push(message);
    }

    let mut app = Model::default();
    app.content = chat;
    app
  }
}

impl Widget for &Message {
  fn render(self, area: Rect, buf: &mut Buffer) {
    let block = Block::bordered().border_set(border::THICK);

    let availible_width: u16 = (area.width as f32 * 0.9).round() as u16;
    let mut new_area = area.clone();
    new_area.width = availible_width;
    //
    // let mut lines: Vec<String> = Vec::new();
    //
    // let mut index = 0;

    Paragraph::new(self.content.clone())
      .wrap(Wrap { trim: true })
      .block(block)
      .render(new_area, buf)
  }
}

fn main() -> color_eyre::Result<()> {
  // tui::install_panic_hook();
  let mut terminal = ratatui::init();
  let mut model = Model::init();

  while model.running_state != RunningState::Done {
    // Render the current view
    terminal.draw(|f| view(&mut model, f))?;

    // Handle events and map to a Message
    let mut current_msg = handle_event(&model)?;

    // Process updates as long as they return a non-None message
    while current_msg.is_some() {
      current_msg = update(&mut model, current_msg.unwrap());
    }
  }

  ratatui::restore();
  Ok(())
}

// TODO: gotta figure out how to model chat state

fn view(model: &mut Model, frame: &mut Frame) {
  let title = Line::from(" Counter App Tutorial ".bold());
  let instructions = Line::from(vec![
    " Decrement ".into(),
    "<Left>".blue().bold(),
    " Increment ".into(),
    "<Right>".blue().bold(),
    " Quit ".into(),
    "<Q> ".blue().bold(),
  ]);
  let block = Block::bordered()
    .title(title.centered())
    .title_bottom(instructions.centered())
    .border_set(border::THICK);

  let _counter_text = Text::from(vec![Line::from(vec![
    "Value: ".into(),
    model.counter.to_string().yellow(),
  ])]);

  let mut messages: Vec<ratatui::prelude::Line> = Vec::new();

  for message in model.content.data.clone() {
    messages.push(Line::from(message.content));
  }

  let message_text = Text::from(messages);

  frame.render_widget(
    Paragraph::new(message_text).right_aligned().block(block),
    frame.area(),
  );
}

/// Convert Event to Action
///
/// We don't need to pass in a `model` to this function in this example
/// but you might need it as your project evolves
fn handle_event(_: &Model) -> color_eyre::Result<Option<Action>> {
  if event::poll(Duration::from_millis(250))? {
    if let Event::Key(key) = event::read()? {
      if key.kind == event::KeyEventKind::Press {
        return Ok(handle_key(key));
      }
    }
  }
  Ok(None)
}

fn handle_key(key: event::KeyEvent) -> Option<Action> {
  match key.code {
    KeyCode::Char('j') => Some(Action::Increment),
    KeyCode::Char('k') => Some(Action::Decrement),
    KeyCode::Char('q') => Some(Action::Quit),
    _ => None,
  }
}

fn update(model: &mut Model, msg: Action) -> Option<Action> {
  match msg {
    Action::Increment => {
      model.counter += 1;
      if model.counter > 50 {
        return Some(Action::Reset);
      }
    }
    Action::Decrement => {
      model.counter -= 1;
      if model.counter < -50 {
        return Some(Action::Reset);
      }
    }
    Action::Reset => model.counter = 0,
    Action::Quit => {
      // You can handle cleanup and exit here
      model.running_state = RunningState::Done;
    }
  };
  None
}
